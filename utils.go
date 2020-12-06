package main

import "github.com/op/go-logging"

func strSliceContains(slc []string, needle string) bool {
	for _, v := range slc {
		if v == needle {
			return true
		}
	}

	return false
}

var log = logging.MustGetLogger("base")