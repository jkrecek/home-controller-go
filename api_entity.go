package main

import (
	"errors"
	"net"
)

type ApiWakePayload struct {
	Mac HwAddress `json:"mac"`

	BroadcastAddress []*ApiBroadcastAddress `json:"addresses,omitempty"`
}

func (w * ApiWakePayload) Validate() error {
	toValidate:= []Validation{w.Mac}

	for _, a := range w.BroadcastAddress {
		toValidate = append(toValidate, a)
	}

	for _, o := range toValidate {
		if err := o.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type ApiBroadcastAddress struct {
	Ip   IP  `json:"ip"`
	Port int `json:"port"`
}

func (a * ApiBroadcastAddress) Validate() error {
	return a.Ip.Validate()
}


type IP string

func (i IP) Validate() error {
	ip := net.ParseIP(string(i))
	if ip == nil {
		return errors.New("invalid ip")
	}

	return nil
}

type HwAddress string

func (a HwAddress) Validate() error {
	_, err := net.ParseMAC(string(a))
	return err
}

type Validation interface {
	Validate() error
}