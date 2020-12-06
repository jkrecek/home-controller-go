package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
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

func parseHaltPayload(r *http.Request) (*ApiHaltPayload, error) {
	allowedJsonMimeTypes := []string{"application/json"}

	contentTypeHeaderValue := r.Header.Get("Content-Type")
	mimeType, _, err := mime.ParseMediaType(contentTypeHeaderValue)
	if err != nil || !strSliceContains(allowedJsonMimeTypes, mimeType) {
		return nil, badRequestError{errors.New("invalid content-type")}
	}

	var haltPayload ApiHaltPayload
	err = json.NewDecoder(r.Body).Decode(&haltPayload)
	defer r.Body.Close()

	if err != nil {
		return nil, badRequestError{errors.New("invalid body")}
	}

	err = haltPayload.Validate()
	if err != nil {
		return nil, badRequestError{fmt.Errorf("invalid body, error: %v", err)}
	}

	return &haltPayload, nil
}

func requireParam(r *http.Request, name string) (string, error) {
	params := mux.Vars(r)
	if uid, ok := params[name]; ok {
		return uid, nil
	}

	return "", badRequestError{fmt.Errorf("invalid param `%s`", name)}
}
