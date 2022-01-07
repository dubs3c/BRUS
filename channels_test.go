package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPreparePayload(t *testing.T) {

	message := "msg"
	textField := "content"
	additionalData := "{\"someKey\":\"someValue\",\"someOtherKey\":\"someOtherValue\"}"
	json, err := PreparePayload(message, textField, additionalData)
	if err != nil {
		t.Fatal("PreparePayload returned error: ", err)
	}
	strJSON := string(json)
	expectedValue := `{"content":"msg","someKey":"someValue","someOtherKey":"someOtherValue"}`
	if strJSON != expectedValue {
		t.Errorf("Expected %s, got %s", expectedValue, strJSON)
	}
}

func TestSendRequest(t *testing.T) {
	// 1. start web server

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	}))
	defer ts.Close()

	_, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	webhook := ts.URL
	json := []byte(`{"content":"msg","someKey":"someValue","someOtherKey":"someOtherValue"}`)
	err = SendRequest(webhook, json)

	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Error("SendRequest returned error: ", err)
	}

}
