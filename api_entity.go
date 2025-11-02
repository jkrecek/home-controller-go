package main

import (
	"errors"
)

type Validation interface {
	Validate() error
}

type ApiWakePayload struct {
	Mac HwAddress `json:"mac"`

	BroadcastAddress []*BroadcastAddress `json:"addresses,omitempty"`
}

func (w *ApiWakePayload) Validate() error {
	toValidate := []Validation{w.Mac}

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

func (w *ApiWakePayload) GetMac() string {
	return string(w.Mac)
}

func (w *ApiWakePayload) GetBroadcastAddress() []*BroadcastAddress {
	return w.BroadcastAddress
}

type ApiHaltPayload struct {
	User       string               `json:"user,required"`
	Host       string               `json:"host,required"`
	Port       *int                 `json:"port,omitempty"`
	Password   Password             `json:"password,omitempty"`
	PrivateKey SshPrivateKeyOptions `json:"private_key,omitempty"`
}

func (h *ApiHaltPayload) Validate() error {
	if len(h.User) == 0 {
		return errors.New("user must not be empty")
	}
	if len(h.Host) == 0 {
		return errors.New("host must not be empty")
	}

	var authErrors []error
	if err := h.Password.Validate(); err != nil {
		authErrors = append(authErrors, err)
	}

	if err := h.PrivateKey.Validate(); err != nil {
		authErrors = append(authErrors, err)
	}

	if len(authErrors) >= 2 {
		return errors.New("either specify password or private_key")
	}

	return nil
}

type ApiStatusData struct {
	IsOnline bool `json:"is_online"`
}
