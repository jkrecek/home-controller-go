package main

import "fmt"

func printStatusResponse(targetConfigId string, isOnline bool) {
	if isOnline {
		fmt.Printf("Target '%s' is ONLINE\n", targetConfigId)
	} else {
		fmt.Printf("Target '%s' is OFFLINE\n", targetConfigId)
	}
}
