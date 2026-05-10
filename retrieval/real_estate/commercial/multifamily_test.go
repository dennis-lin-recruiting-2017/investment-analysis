package commercial

import (
	"fmt"
	"investment-analysis/llm"
	"investment-analysis/persistence/model"
	"investment-analysis/persistence/sqlite"
	"investment-analysis/retrieval"
	"investment-analysis/util"
	"testing"
)

func TestGetMultifamily01(t *testing.T) {
	store, err := sqlite.NewStore("./my-docs.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	client, err := retrieval.NewPlaywrightClient("", store.Documents, store.RetrievalAttempts, 300)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	key := "test-commercial-real-estate"
	url := "https://www.loopnet.com/Listing/525-Hamilton-Ave-Palo-Alto-CA/40362864/"
	output_label := "listing -- commercial real estate"

	err = client.SaveDocument(util.NewTraceContext(), url, key, model.ListingCommercial, output_label)
	if err != nil {
		t.Fatal(err)
	}

	document, err := store.Documents.GetBody(util.NewTraceContext(), key)
	if err != nil {
		t.Fatal(err)
	}
	//print(string(document))
	question := `Replace the value of the JSON object with the answer to the respective question
  {
    "property_type": "What type of property is this -- multi-family, retail, office, hospitality, or manufacturing?",
    "address": {
      "street_number": "What is the street number?",
      "street_name": "What is the street name?",
      "street_type": "What is the street type?",
      "county": "What is the name of the county this property is in?",
      "city": "What is the name of the city this property is in?",
      "state": "What is the name of the state this property is in?",
      "postal_code": "What is the postal code that this property is in?",
    },
    "documents": "Create an array of hyperlinks of documents that can be downloaded.  IF there are no documents, then create an empty array"
  }

  Return this JSON object as the response.
  `
	print(document)
	print(question)

	existingRecord, err := store.Inferences.Get(util.NewTraceContext(), "test-inference-123")

	llmClient := llm.NewClient(model.DefaultSettings(""))
	messages := make([]llm.Message, 1)
	messages[0] = llm.Message{
		Role:    "user",
		Content: fmt.Sprintf("%s\n\n%s", document, question),
		//Content: "Fried chicken recipe",
	}
	response, err := llmClient.Chat(util.NewTraceContext(), messages)
	if err != nil {
		t.Fatal(err)
	}
	print(response.Content)
	t.Logf("usage: prompt=%d completion=%d total=%d, elapsed=%s",
		response.PromptTokens, response.CompletionTokens, response.TotalTokens, response.Elapsed)
}
