package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func handleRemoteCommand(remoteId, targetId string, command string) {
	if remoteId == "" {
		log.Fatal("missing flag --remote")
		return
	}

	if targetId == "" {
		log.Fatal("missing flag --target")
		return
	}

	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	remoteConfig := getRemoteConfigurationById(config, remoteId)

	if remoteConfig == nil {
		log.Fatalf("Configuration '%s' not found in config file", remoteId)
		return
	}

	targetConfig := getTargetConfigurationById(&remoteConfig.Targets, targetId)
	if targetConfig == nil {
		log.Fatalf("Target '%s' not found in for configuration %s", targetId, remoteConfig.Id)
		return
	}

	requestOpts, responseOpts, err := getRequestOpts(targetConfig, command)
	if err != nil {
		log.Fatalf("Cannot handle command '%s': %v", command, err)
		return
	}

	fullUrl, err := url.JoinPath(remoteConfig.Host, requestOpts.Path)
	if err != nil {
		log.Fatalf("Cannot build url for request: %v", err)
		return
	}

	var reader io.Reader = nil
	if requestOpts.Body != nil {
		body, err := json.Marshal(requestOpts.Body)
		if err != nil {
			log.Fatalf("Cannot marshal body: %v", err)
			return
		}

		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(requestOpts.Method, fullUrl, reader)
	if err != nil {
		log.Fatalf("Request processing failed: %v", err)
		return
	}

	if remoteConfig.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", remoteConfig.AuthToken))
	}

	if reader != nil {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
		return
	}

	if resp.StatusCode >= 400 {
		log.Fatalf("Request failed with status code %d", resp.StatusCode)
		return
	}

	defer resp.Body.Close()

	err = responseOpts.OnSuccess(resp)
	if err != nil {
		log.Fatalf("Response processing failed: %v", err)
		return
	}
}

type RequestOpts struct {
	Method string
	Path   string
	Body   interface{}
}

type ResponseOpts struct {
	OnSuccess func(response *http.Response) error
}

func getRequestOpts(targetConfig *TargetConfiguration, command string) (*RequestOpts, *ResponseOpts, error) {
	switch command {
	case "wake":
		return getRemoteWakeRequestOpts(targetConfig)
	case "halt":
		return getRemoteHaltRequestOpts(targetConfig)
	case "status":
		return getRemoteStatusRequestOpts(targetConfig)
	//case "status-stream":
	default:
		return nil, nil, fmt.Errorf("unknown command '%s'", command)
	}
}

func getRemoteWakeRequestOpts(targetConfig *TargetConfiguration) (*RequestOpts, *ResponseOpts, error) {
	body := &ApiWakePayload{
		Mac:              HwAddress(targetConfig.GetMac()),
		BroadcastAddress: targetConfig.GetBroadcastAddress(),
	}
	requestOpts := &RequestOpts{
		Method: "POST",
		Path:   "/wake",
		Body:   body,
	}

	successResponseHandler := func(response *http.Response) error {
		fmt.Printf("Wake request sent to %s.\n", targetConfig.Host)
		return nil
	}

	responseOpts := &ResponseOpts{
		OnSuccess: successResponseHandler,
	}
	return requestOpts, responseOpts, nil
}

func getRemoteHaltRequestOpts(targetConfig *TargetConfiguration) (*RequestOpts, *ResponseOpts, error) {
	body := &ApiHaltPayload{
		User:       targetConfig.Ssh.User,
		Host:       targetConfig.Host,
		Port:       targetConfig.Ssh.Port,
		Password:   targetConfig.Ssh.Password,
		PrivateKey: targetConfig.Ssh.PrivateKey,
	}
	requestOpts := &RequestOpts{
		Method: "POST",
		Path:   "/halt",
		Body:   body,
	}

	successResponseHandler := func(response *http.Response) error {
		fmt.Printf("Halt request sent to %s.\n", targetConfig.Host)
		return nil
	}

	responseOpts := &ResponseOpts{
		OnSuccess: successResponseHandler,
	}

	return requestOpts, responseOpts, nil
}

func getRemoteStatusRequestOpts(targetConfig *TargetConfiguration) (*RequestOpts, *ResponseOpts, error) {
	requestOpts := &RequestOpts{
		Method: "GET",
		Path:   fmt.Sprintf("/status/%s", targetConfig.Host),
	}

	successResponseHandler := func(response *http.Response) error {
		var status ApiStatusData
		err := decodeResponseBody(response, &status)
		if err != nil {
			log.Fatalf("Cannot decode response body: %v", err)
			return nil
		}

		printStatusResponse(targetConfig.Id, status.IsOnline)
		return nil
	}

	responseOpts := &ResponseOpts{
		OnSuccess: successResponseHandler,
	}
	return requestOpts, responseOpts, nil
}
