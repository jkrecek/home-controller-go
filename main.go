package main

import (
	"flag"
)

var (
	syncQuit = make(chan struct{})
)

var httpAddrFlag = flag.String("http_addr", ":80", "Address to which HTTP server should bind")
var httpsAddrFlag = flag.String("https_addr", ":443", "Address to which HTTPS server should bind")
var httpsCertFlag = flag.String("https_cert", "", "Path to file containing HTTPS certificate")
var httpsKeyFlag = flag.String("https_key", "", "Path to file containing HTTPS key")

func main() {
	flag.Parse()

	api := InitApiCore()
	api.SetHttp(*httpAddrFlag)
	api.SetHttps(*httpsAddrFlag, *httpsCertFlag, *httpsKeyFlag)
	api.StartListen()

	<-syncQuit
}
