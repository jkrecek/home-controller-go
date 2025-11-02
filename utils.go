package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/op/go-logging"
	"golang.org/x/term"
)

func strSliceContains(slc []string, needle string) bool {
	for _, v := range slc {
		if v == needle {
			return true
		}
	}

	return false
}

var log = logging.MustGetLogger("base")

func readPassword() (string, error) {
	var rawInput string
	if term.IsTerminal(int(os.Stdin.Fd())) {
		pwBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println() // ReadPassword doesn't print newline
		if err != nil {
			return "", err
		}

		rawInput = string(pwBytes)

	} else {
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		rawInput = input
	}

	return strings.TrimSpace(rawInput), nil
}
