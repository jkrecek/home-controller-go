package main

import "fmt"

func handleRunCommand(args []string) {
	if len(args) < 3 {
		log.Fatal("command run must have at least 2 arguments: homecontroller run command [target]")
		return
	}

	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	targetId := args[2]
	targetConfig := getTargetConfigurationById(&config.RunTargets, targetId)
	if targetConfig == nil {
		log.Fatalf("Run target '%s' not found", targetId)
		return
	}

	switch args[1] {
	case "wake":
		handleRunWake(targetConfig)
		break
	case "halt":
		handleRunHalt(targetConfig)
		break
	case "status":
		handleRunStatus(targetConfig)
		break
	case "status-stream":
	default:
		log.Fatalf("Unknown command '%s'", args[2])
		break
	}
}

func handleRunWake(targetConfig *TargetConfiguration) {
	sendMagicPacket(targetConfig)

	fmt.Printf("Magic packet sent to '%s' to mac '%s'\n", targetConfig.Id, targetConfig.Mac)
}

func handleRunHalt(targetConfig *TargetConfiguration) {
	_, err := haltViaSsh(targetConfig.Ssh.User, targetConfig.Host, targetConfig.Ssh.Port, targetConfig.Ssh.Password, &targetConfig.Ssh.PrivateKey, func() string {
		fmt.Print("Enter passphrase for private key: ")
		pwd, err := readPassword()
		if err != nil {
			log.Fatalf("Could not read password: %v", err)
		}
		return pwd
	})
	if err != nil {
		log.Errorf("Could not send halt command via ssh to target %s: %v", targetConfig.Id, err)
	}

	fmt.Printf("Halt command sent to '%s'\n", targetConfig.Id)
}

func handleRunStatus(targetConfig *TargetConfiguration) {
	isOnline, err := pingToCheckOnline(targetConfig.Host)
	if err != nil {
		log.Errorf("Could not check online status of target %s: %v", targetConfig.Id, err)
	}

	printStatusResponse(targetConfig.Id, isOnline)
}
