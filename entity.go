package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
	"path"

	"golang.org/x/crypto/ssh"
)

type BroadcastAddress struct {
	Ip   IP  `json:"ip" yaml:"ip"`
	Port int `json:"port" yaml:"port"`
}

func (a *BroadcastAddress) Validate() error {
	return a.Ip.Validate()
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

type HwAddress string

func (a HwAddress) Validate() error {
	_, err := net.ParseMAC(string(a))
	return err
}

type IP string

func (i IP) Validate() error {
	ip := net.ParseIP(string(i))
	if ip == nil {
		return errors.New("invalid ip")
	}

	return nil
}

type SshPrivateKeyOptions struct {
	Path       string `json:"path,required" yaml:"path"`
	Passphrase string `json:"passphrase" yaml:"passphrase"`
}

func (k *SshPrivateKeyOptions) Validate() error {
	if len(k.Path) == 0 {
		return errors.New("path must not be empty")
	}

	return nil
}

func (k *SshPrivateKeyOptions) GetFullPath() (string, error) {
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

func (k *SshPrivateKeyOptions) AuthMethod(requestPassphrase func() string) (*ssh.AuthMethod, error) {
	if k == nil {
		return nil, errors.New("missing private key")
	}

	keyPath, err := k.GetFullPath()
	if err != nil {
		return nil, err
	}

	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	// Create the Signer for this private key.
	signer, err := sshAuthSigner(key, k.Passphrase, requestPassphrase)
	if err != nil {
		return nil, err
	}

	authMethod := ssh.PublicKeys(signer)
	return &authMethod, nil
}
