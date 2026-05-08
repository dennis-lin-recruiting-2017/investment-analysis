package commercial

import (
	"context"
	"investment-analysis/persistence"
	"investment-analysis/retrieval"
	"testing"
)

func TestGetMultifamily01(t *testing.T) {
	store, err := persistence.OpenFile("./my-docs.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	client, err := retrieval.NewPlaywrightClient("", store, 300)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	key := "test-commercial-real-estate"
	document_type := "listing-commercial"
	url := "https://www.loopnet.com/Listing/525-Hamilton-Ave-Palo-Alto-CA/40362864/"
	output_label := "listing -- commercial real estate"

	err = client.SaveDocument(context.Background(), url, key, document_type, output_label)
	if err != nil {
		t.Fatal(err)
	}

	document, err := store.GetDocumentBody(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	print(string(document))
}
