package main

import (
	"net/http"
)

func (h *httpApiHandler) Wake(r *http.Request) (interface{}, error) {
	wakePayload, err := parseWakePayload(r)
	if err != nil {
		return nil, err
	}

	sendMagicPacket(wakePayload)

	// sends magic packet
	return nil, nil
}

func (h *httpApiHandler) Halt(r *http.Request) (interface{}, error) {
	haltPayload, err := parseHaltPayload(r)
	if err != nil {
		return nil, err
	}

	shouldSudo := haltPayload.User != "root"
	cmd := "halt -p"
	if shouldSudo {
		cmd = "sudo " + cmd
	}

	_, err = openSshSessionCommand(haltPayload.User, haltPayload.Host, haltPayload.GetPort(), haltPayload.Password, &haltPayload.PrivateKey, cmd)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *httpApiHandler) Status(r *http.Request) (interface{}, error) {
	host, err := requirePathParam(r, "host")
	if err != nil {
		return nil, err
	}

	isOnline, err := pingToCheckOnline(host)
	if err != nil {
		return nil, err
	}

	statusData := ApiStatusData{
		IsOnline: isOnline,
	}

	// starts pinging service
	// run 5 times, if at least one is ok
	return statusData, nil
}
