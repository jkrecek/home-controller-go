package main

import (
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

const userRelativeConfigPath = ".homecontroller/config.yml"

func getPrintConfigPath() string {
	return fmt.Sprintf("~/%s", userRelativeConfigPath)
}

type SshConfiguration struct {
	User       string               `yaml:"user"`
	Port       *int                 `yaml:"port"`
	Password   Password             `yaml:"password,omitempty"`
	PrivateKey SshPrivateKeyOptions `yaml:"private_key,omitempty"`
}

type TargetConfiguration struct {
	Id               string              `yaml:"id"`
	Host             string              `yaml:"host"`
	Mac              HwAddress           `yaml:"mac"`
	Ssh              SshConfiguration    `yaml:"ssh"`
	BroadcastAddress []*BroadcastAddress `yaml:"broadcast_address,omitempty"`
}

func (t *TargetConfiguration) GetMac() string {
	return string(t.Mac)
}

func (t *TargetConfiguration) GetBroadcastAddress() []*BroadcastAddress {
	return t.BroadcastAddress
}

type RemoteConfiguration struct {
	Id        string                `yaml:"id"`
	Host      string                `yaml:"host"`
	AuthToken string                `yaml:"auth_token"`
	Targets   []TargetConfiguration `yaml:"targets"`
}

type LocalConfiguration struct {
	RunTargets []TargetConfiguration `yaml:"run_targets"`
	Remote     []RemoteConfiguration `yaml:"remote"`
}

func loadConfig() (*LocalConfiguration, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	bts, err := os.ReadFile(path.Join(dirname, userRelativeConfigPath))
	if err != nil {
		return nil, fmt.Errorf("couldnt read config file '%s'", getPrintConfigPath())
	}

	var config LocalConfiguration
	err = yaml.Unmarshal(bts, &config)
	if err != nil {
		return nil, fmt.Errorf("couldnt parse config file '%s' %v", getPrintConfigPath(), err)
	}

	return &config, nil
}

func getRemoteConfigurationById(config *LocalConfiguration, id string) *RemoteConfiguration {
	for i := range config.Remote {
		if config.Remote[i].Id == id {
			return &config.Remote[i]
		}
	}
	return nil
}

func getTargetConfigurationById(targets *[]TargetConfiguration, id string) *TargetConfiguration {
	for i := range *targets {
		if (*targets)[i].Id == id {
			return &(*targets)[i]
		}
	}
	return nil
}
