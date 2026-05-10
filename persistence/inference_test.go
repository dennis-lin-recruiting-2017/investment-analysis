package persistence_test

import (
	"errors"
	"investment-analysis/util"
	"path/filepath"
	"testing"

	"investment-analysis/persistence"
	"investment-analysis/persistence/model"
	"investment-analysis/persistence/sqlite"
)

func newTestStore(t *testing.T) *persistence.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := sqlite.NewStore(dbPath)
	if err != nil {
		t.Fatalf("sqlite.NewStore: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestSaveInference_AssignsUUIDAndPersistsMetadata(t *testing.T) {
	store := newTestStore(t)
	ctx := util.NewTraceContext()

	in := model.Inference{
		InvestmentUUID:   "inv-123",
		Model:            "test-model",
		Prompt:           "what is 2+2?",
		Response:         "4",
		PromptTokens:     11,
		CompletionTokens: 1,
		TotalTokens:      12,
		RequestedAt:      "2026-05-09T10:00:00Z",
		RespondedAt:      "2026-05-09T10:00:01Z",
		ElapsedMs:        1000,
	}

	saved, err := store.Inferences.Save(ctx, in)
	if err != nil {
		t.Fatalf("Inferences.Save: %v", err)
	}
	if saved.UUID == "" {
		t.Errorf("expected Save to assign a UUID, got empty")
	}
	if saved.InvestmentUUID != "inv-123" || saved.Prompt != "what is 2+2?" || saved.Response != "4" {
		t.Errorf("saved fields not preserved: %+v", saved)
	}
	if saved.PromptTokens != 11 || saved.CompletionTokens != 1 || saved.TotalTokens != 12 {
		t.Errorf("token usage not preserved: %+v", saved)
	}
	if saved.RequestedAt != "2026-05-09T10:00:00Z" || saved.RespondedAt != "2026-05-09T10:00:01Z" || saved.ElapsedMs != 1000 {
		t.Errorf("timing metadata not preserved: %+v", saved)
	}
}

func TestSaveInference_RejectsMissingInvestmentUUID(t *testing.T) {
	store := newTestStore(t)
	_, err := store.Inferences.Save(util.NewTraceContext(), model.Inference{Prompt: "x"})
	if err == nil {
		t.Fatal("expected error when InvestmentUUID is empty, got nil")
	}
}

func TestGetInference_RoundTrip(t *testing.T) {
	store := newTestStore(t)
	ctx := util.NewTraceContext()
	saved, err := store.Inferences.Save(ctx, model.Inference{
		InvestmentUUID: "inv-123",
		Prompt:         "ping",
		Response:       "pong",
	})
	if err != nil {
		t.Fatalf("Inferences.Save: %v", err)
	}

	got, err := store.Inferences.Get(ctx, saved.UUID)
	if err != nil {
		t.Fatalf("Inferences.Get: %v", err)
	}
	if got.UUID != saved.UUID || got.Response != "pong" {
		t.Errorf("Get round-trip: got %+v want %+v", got, saved)
	}

	if _, err := store.Inferences.Get(ctx, "does-not-exist"); !errors.Is(err, model.ErrNotFound) {
		t.Errorf("missing UUID should return ErrNotFound, got %v", err)
	}
}

func TestListInferencesByInvestment_OrderedNewestFirst(t *testing.T) {
	store := newTestStore(t)
	ctx := util.NewTraceContext()

	for _, prompt := range []string{"first", "second", "third"} {
		if _, err := store.Inferences.Save(ctx, model.Inference{
			InvestmentUUID: "inv-A",
			Prompt:         prompt,
		}); err != nil {
			t.Fatalf("Inferences.Save %q: %v", prompt, err)
		}
	}
	if _, err := store.Inferences.Save(ctx, model.Inference{InvestmentUUID: "inv-B", Prompt: "other"}); err != nil {
		t.Fatalf("Inferences.Save inv-B: %v", err)
	}

	got, err := store.Inferences.ListByInvestment(ctx, "inv-A")
	if err != nil {
		t.Fatalf("Inferences.ListByInvestment: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d rows for inv-A; want 3", len(got))
	}
	want := []string{"third", "second", "first"}
	for i, w := range want {
		if got[i].Prompt != w {
			t.Errorf("row[%d].Prompt = %q; want %q", i, got[i].Prompt, w)
		}
	}
}

func TestGetByKey_RoundTripAndScoping(t *testing.T) {
	store := newTestStore(t)
	ctx := util.NewTraceContext()

	// Save two inferences for inv-A under distinct keys, plus a same-key
	// row under inv-B to verify the lookup is scoped to the investment.
	if _, err := store.Inferences.Save(ctx, model.Inference{
		InvestmentUUID: "inv-A",
		Key:            "summary-2026-q1",
		Prompt:         "summarise the q1 filing",
	}); err != nil {
		t.Fatalf("Save inv-A summary: %v", err)
	}
	if _, err := store.Inferences.Save(ctx, model.Inference{
		InvestmentUUID: "inv-A",
		Key:            "thesis-2026-q1",
		Prompt:         "score the thesis",
	}); err != nil {
		t.Fatalf("Save inv-A thesis: %v", err)
	}
	if _, err := store.Inferences.Save(ctx, model.Inference{
		InvestmentUUID: "inv-B",
		Key:            "summary-2026-q1",
		Prompt:         "different investment",
	}); err != nil {
		t.Fatalf("Save inv-B summary: %v", err)
	}

	got, err := store.Inferences.GetByKey(ctx, "inv-A", "summary-2026-q1")
	if err != nil {
		t.Fatalf("GetByKey: %v", err)
	}
	if got.InvestmentUUID != "inv-A" || got.Key != "summary-2026-q1" || got.Prompt != "summarise the q1 filing" {
		t.Errorf("GetByKey returned wrong row: %+v", got)
	}

	// Missing key on a known investment → ErrNotFound.
	if _, err := store.Inferences.GetByKey(ctx, "inv-A", "no-such-key"); !errors.Is(err, model.ErrNotFound) {
		t.Errorf("GetByKey missing: got %v; want ErrNotFound", err)
	}

	// Same key under a different investment is independently retrievable.
	other, err := store.Inferences.GetByKey(ctx, "inv-B", "summary-2026-q1")
	if err != nil {
		t.Fatalf("GetByKey inv-B: %v", err)
	}
	if other.InvestmentUUID != "inv-B" || other.Prompt != "different investment" {
		t.Errorf("GetByKey inv-B returned wrong row: %+v", other)
	}
}

func TestUpdateInferenceArtifact(t *testing.T) {
	store := newTestStore(t)
	ctx := util.NewTraceContext()
	saved, err := store.Inferences.Save(ctx, model.Inference{
		InvestmentUUID: "inv-123",
		Response:       `{"raw":"data"}`,
	})
	if err != nil {
		t.Fatalf("Inferences.Save: %v", err)
	}

	if err := store.Inferences.UpdateArtifact(ctx, saved.UUID, "extracted-value"); err != nil {
		t.Fatalf("Inferences.UpdateArtifact: %v", err)
	}

	reloaded, err := store.Inferences.Get(ctx, saved.UUID)
	if err != nil {
		t.Fatalf("Inferences.Get: %v", err)
	}
	if reloaded.ProcessedArtifact != "extracted-value" {
		t.Errorf("ProcessedArtifact = %q; want %q", reloaded.ProcessedArtifact, "extracted-value")
	}

	if err := store.Inferences.UpdateArtifact(ctx, "missing", "x"); !errors.Is(err, model.ErrNotFound) {
		t.Errorf("UpdateArtifact on missing UUID: got %v; want ErrNotFound", err)
	}
}
