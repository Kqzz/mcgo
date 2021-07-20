package mcgo

import (
	"net/http"
	"testing"
)

func TestGenPayload(t *testing.T) {
	header := http.Header{
		"Authorization": {"Hi"},
	}

	validPayload := "POST /signup HTTP/1.1\r\nHost: example.com\r\nAuthorization: Hi\r\n"
	payload, err := generatePayload("POST", "http://example.com/signup", header, "")

	if err != nil || payload != validPayload {
		t.Fatalf("err: %v | payload: %v | expected payload: %v", err, payload, validPayload)
	}
}
