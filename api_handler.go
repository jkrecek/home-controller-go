package main

import (
	"github.com/linde12/gowol"
	"net/http"
	"strconv"
)

func (h *httpApiHandler) Wake(r *http.Request) (interface{}, error) {

	wakePayload, err := parseWakePayload(r)
	if err != nil {
		return nil, err
	}

	if packet, err := gowol.NewMagicPacket(string(wakePayload.Mac)); err == nil {
		if len(wakePayload.BroadcastAddress) != 0 {
			for i := 0;i < len(wakePayload.BroadcastAddress) - 1 ;i++ {
				address :=wakePayload.BroadcastAddress[i]
				packet.SendPort(string(address.Ip), strconv.Itoa(address.Port))
			}
		} else {
			defaultPorts := []int{7,9}
			for i := 0;i < len(defaultPorts) - 1 ;i++ {
				packet.SendPort("255.255.255.255", strconv.Itoa(defaultPorts[i]))
			}
		}
	}

	// sends magic packet
	return nil, nil
}

func (h* httpApiHandler) Halt(r *http.Request) (interface{}, error) {

	// ssh host 'sudo halt -p'
	return nil, nil
}

func (h* httpApiHandler) IsOnline(r *http.Request) (interface{}, error) {

	// starts pinging service
	// run 5 times, if at least one is ok
	return nil, nil
}

