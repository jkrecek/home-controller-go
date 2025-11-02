package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	syncQuit = make(chan struct{})
)

var httpAuthTokenFlag = flag.String("auth_token", "", "Token that must be provided in HTTP header to access API")
var httpAddrFlag = flag.String("http_addr", ":80", "Address to which HTTP server should bind")
var httpsAddrFlag = flag.String("https_addr", ":443", "Address to which HTTPS server should bind")
var httpsCertFlag = flag.String("https_cert", "", "Path to file containing HTTPS certificate")
var httpsKeyFlag = flag.String("https_key", "", "Path to file containing HTTPS key")
var cmdTargetFlag = flag.String("target", "", "Identifier of target in config to run command for")
var cmdRemoteFlag = flag.String("remote", "", "Identifier of remote server, via which commands should run")

func failWithUsage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	fmt.Fprintln(flag.CommandLine.Output(), " Commands:")
	fmt.Fprintln(flag.CommandLine.Output(), "  http: Start HTTP server")
	fmt.Fprintln(flag.CommandLine.Output(), "  run: Runs command directly")
	fmt.Fprintln(flag.CommandLine.Output(), "  remote-run: Runs command via remote server")
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

		if *httpAuthTokenFlag != "" {
			api.UseMiddleware(createTokenAuthMiddleware(*httpAuthTokenFlag))
		}
		api.StartListen()

		<-syncQuit
		break
	case "run":
		if len(args) < 2 {
			log.Fatal("command run must have an argument: homecontroller --target=[target] run [COMMAND]")
			return
		}

		handleRunCommand(*cmdTargetFlag, args[1])
		break

	case "remote-run":
		if len(args) < 2 {
			log.Fatal("command remote-run must have an argument: homecontroller --remote=[remote] --target=[target] remote-run [COMMAND]")
			return
		}
		handleRemoteCommand(*cmdRemoteFlag, *cmdTargetFlag, args[1])
		break
	default:
		failWithUsage()
		break
	}

}
