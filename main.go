package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	syncQuit = make(chan struct{})
)

var httpAddrFlag = flag.String("http_addr", ":80", "Address to which HTTP server should bind")
var httpsAddrFlag = flag.String("https_addr", ":443", "Address to which HTTPS server should bind")
var httpsCertFlag = flag.String("https_cert", "", "Path to file containing HTTPS certificate")
var httpsKeyFlag = flag.String("https_key", "", "Path to file containing HTTPS key")

func failWithUsage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	fmt.Fprintln(flag.CommandLine.Output(), " Commands:")
	fmt.Fprintln(flag.CommandLine.Output(), "  http: Start HTTP server")
	fmt.Fprintln(flag.CommandLine.Output(), "  run: Runs command directly")
	fmt.Fprintln(flag.CommandLine.Output(), "  remotely: Runs command via remote server")
	fmt.Fprintln(flag.CommandLine.Output(), "")
	fmt.Fprintln(flag.CommandLine.Output(), " Flags:")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		failWithUsage()
	}

	switch args[0] {
	case "http":
		api := InitApiCore()
		api.SetHttp(*httpAddrFlag)
		api.SetHttps(*httpsAddrFlag, *httpsCertFlag, *httpsKeyFlag)
		api.StartListen()

		<-syncQuit
		break
	case "run":
		handleRunCommand(args)
		break

	case "remote":
		handleRemoteCommand(args)
		break
	default:
		failWithUsage()
		break
	}

}
