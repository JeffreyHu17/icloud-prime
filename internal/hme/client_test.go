package hme

import (
	"encoding/json"
	"testing"
)

func TestMailDomainRequestsUseJSONHeaders(t *testing.T) {
	contentType, acceptType := requestTypes("p68-maildomainws.icloud.com")

	if contentType != "application/json" {
		t.Fatalf("expected application/json content type, got %q", contentType)
	}
	if acceptType != "application/json" {
		t.Fatalf("expected application/json accept type, got %q", acceptType)
	}
}

func TestGeneratePayloadIsEmptyObject(t *testing.T) {
	raw, err := json.Marshal(generatePayload())
	if err != nil {
		t.Fatal(err)
	}

	if string(raw) != "{}" {
		t.Fatalf("expected empty JSON object payload, got %s", raw)
	}
}
