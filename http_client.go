package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
)

func decodeResponseBody(response *http.Response, v any) error {
	defer response.Body.Close()
	contentType := response.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return err
	}

	if mediaType != "application/json" {
		return fmt.Errorf("unexpected content type '%s', expected: 'application/json'", contentType)
	}

	decoder := json.NewDecoder(response.Body)
	return decoder.Decode(v)
}
