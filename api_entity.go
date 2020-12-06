package main

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"os/user"
	"path"
)

type Validation interface {
	Validate() error
}

type ApiWakePayload struct {
	Mac HwAddress `json:"mac"`

	BroadcastAddress []*ApiBroadcastAddress `json:"addresses,omitempty"`
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

type ApiBroadcastAddress struct {
	Ip   IP  `json:"ip"`
	Port int `json:"port"`
}

func (a *ApiBroadcastAddress) Validate() error {
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

type ApiHaltPayload struct {
	User       string           `json:"user,required"`
	Host       string           `json:"host,required"`
	Port       *int             `json:"port,omitempty"`
	Password   Password         `json:"password,omitempty"`
	PrivateKey ApiSshPrivateKey `json:"private_key,omitempty"`
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

func (h *ApiHaltPayload) GetPort() int {
	if h.Port != nil {
		return *h.Port
	} else {
		return 22
	}
}

type Password string

func (p Password) Validate() error {
	if len(p) == 0 {
		return errors.New("password must not be empty")
	}

	return nil
}

func (p Password) AuthMethod() ssh.AuthMethod {
	if len(p) == 0 {
		return nil
	}

	return ssh.Password(string(p))
}

type ApiSshPrivateKey struct {
	Path       string `json:"path,required"`
	Passphrase string `json:"passphrase"`
}

func (k *ApiSshPrivateKey) Validate() error {
	if len(k.Path) == 0 {
		return errors.New("path must not be empty")
	}

	return nil
}

func (k *ApiSshPrivateKey) GetFullPath() (string, error) {
	if path.IsAbs(k.Path) {
		return k.Path, nil
	} else {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}

		fullPath := fmt.Sprintf("%s/.ssh/%s", usr.HomeDir, k.Path)
		return fullPath, nil
	}
}

func (k *ApiSshPrivateKey) AuthMethod() ssh.AuthMethod {
	if k == nil {
		return nil
	}

	keyPath, err := k.GetFullPath()
	if err != nil {
		return nil
	}

	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil
	}

	// Create the Signer for this private key.
	signer, err := sshAuthSigner(key, k.Passphrase)
	if err != nil {
		log.Warning(err)
		return nil
	}

	return ssh.PublicKeys(signer)
}

type ApiStatusData struct {
	IsOnline bool `json:"is_online"`
}
