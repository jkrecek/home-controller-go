package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
)

func parseWakePayload(r *http.Request) (*ApiWakePayload, error) {
	allowedJsonMimeTypes := []string{"application/json"}

	contentTypeHeaderValue := r.Header.Get("Content-Type")
	mimeType, _, err := mime.ParseMediaType(contentTypeHeaderValue)
	if err != nil || !strSliceContains(allowedJsonMimeTypes, mimeType) {
		return nil, badRequestError{errors.New("invalid content-type")}
	}

	var wakePayload ApiWakePayload
	err = json.NewDecoder(r.Body).Decode(&wakePayload)
	defer r.Body.Close()

	if err != nil {
		return nil, badRequestError{errors.New("invalid body")}
	}

	err = wakePayload.Validate()
	if err != nil {
		return nil, badRequestError{fmt.Errorf("invalid body, error: %v", err)}
	}

	return &wakePayload, nil
}
