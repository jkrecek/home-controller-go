package main

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"time"

	"github.com/go-ping/ping"
	"github.com/linde12/gowol"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type ActionMagicPacket interface {
	GetMac() string
	GetBroadcastAddress() []*BroadcastAddress
}

func sendMagicPacket(magicPacket ActionMagicPacket) {
	if packet, err := gowol.NewMagicPacket(magicPacket.GetMac()); err == nil {
		broadcastAddress := magicPacket.GetBroadcastAddress()
		if len(broadcastAddress) != 0 {
			for i := 0; i < len(broadcastAddress)-1; i++ {
				address := broadcastAddress[i]
				packet.SendPort(string(address.Ip), strconv.Itoa(address.Port))
			}
		} else {
			defaultPorts := []int{7, 9}
			for i := 0; i < len(defaultPorts)-1; i++ {
				packet.SendPort("255.255.255.255", strconv.Itoa(defaultPorts[i]))
			}
		}
	}
}

func pingToCheckOnline(host string) (bool, error) {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return false, err
	}

	pinger.Timeout = 3 * time.Second
	pinger.OnRecv = func(packet *ping.Packet) {
		pinger.Stop()
	}

	err = pinger.Run()
	if err != nil {
		return false, err
	}

	stats := pinger.Statistics()
	anySuccess := stats.PacketsRecv > 0

	return anySuccess, nil
}

func sshKnownHosts() (ssh.HostKeyCallback, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("couldnt access user home %s", err)
	}

	path := fmt.Sprintf("%s/.ssh/known_hosts", usr.HomeDir)
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("couldnt access known_hosts file %s", err)
	}
	defer file.Close()

	hostKeyCallback, err := knownhosts.New(path)
	if err != nil {
		return nil, fmt.Errorf("couldnt create new knownhosts %s", err)
	}

	return hostKeyCallback, err
}

func sshAuthSigner(pemKey []byte, passphrase string, requestPassphrase func() string) (ssh.Signer, error) {
	if len(passphrase) > 0 {
		return ssh.ParsePrivateKeyWithPassphrase(pemKey, []byte(passphrase))
	} else {
		signer, err := ssh.ParsePrivateKey(pemKey)
		if err != nil && requestPassphrase != nil {
			if _, ok := err.(*ssh.PassphraseMissingError); ok {
				return sshAuthSigner(pemKey, requestPassphrase(), nil)
			}
		}
		return signer, err
	}
}

func haltViaSsh(user string, host string, port *int, password Password, privateKey *SshPrivateKeyOptions, requestPassphrase func() string) ([]byte, error) {
	shouldSudo := user != "root"
	cmd := "halt -p"
	if shouldSudo {
		cmd = "sudo " + cmd
	}
	return openSshSessionCommand(user, host, port, password, privateKey, cmd, requestPassphrase)
}

func openSshSessionCommand(user string, host string, port *int, password Password, privateKey *SshPrivateKeyOptions, cmd string, requestPassphrase func() string) ([]byte, error) {
	hostKeyCallback, err := sshKnownHosts()
	if err != nil {
		return nil, err
	}

	realPort := 22
	if port != nil {
		realPort = *port
	}

	var authMethods []ssh.AuthMethod
	if privateKey != nil {
		authKey, err := privateKey.AuthMethod(requestPassphrase)
		if err != nil {
			return nil, err
		}

		authMethods = append(authMethods, *authKey)
	}

	authPass := password.AuthMethod()
	if authPass != nil {
		authMethods = append(authMethods, authPass)
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
	}

	addr := fmt.Sprintf("%s:%d", host, realPort)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("couldnt dial ssh, %s", err)
	}

	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("couldnt create client session, %s", err)
	}

	defer session.Close()

	// ignore result from output
	session.Output(cmd)
	return nil, nil
}

func observePingOnHost(host string, done chan bool, update func(status ApiStatusData)) error {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return err
	}

	defer pinger.Stop()

	pinger.RecordRtts = false

	var lastReceived time.Time
	// received := false
	pinger.OnRecv = func(packet *ping.Packet) {
		lastReceived = time.Now()
		// received = true
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				isOnline := time.Now().Before(lastReceived.Add(2 * time.Second))
				update(ApiStatusData{
					IsOnline: isOnline,
				})
			}
		}
	}()

	go func() {
		err = pinger.Run()
		if err != nil {
			log.Error(err)
		}
	}()

	<-done
	return nil
}
