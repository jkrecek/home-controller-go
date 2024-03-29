package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"io"
)

func (h *httpApiHandler) StatusStream(conn *websocket.Conn) {
	host, err := requireQueryParam(conn.Request(), "host")
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("Connection error: %s", err)))
		conn.Close()
		return
	}

	done := make(chan bool)

	go func() {
		var msg = make([]byte, 512)
		if _, err := conn.Read(msg); err != nil {
			if err == io.EOF {
				done <- true
			} else {
				log.Error(err)
			}

		}
	}()

	err = observePingOnHost(host, done, func(status ApiStatusData) {
		b, err := json.Marshal(status)
		if err != nil {
			log.Warning(err)
			return
		}

		_, err = conn.Write(b)
		if err != nil {
			done <- true
		}
	})

	if err != nil {
		log.Error(err)
	}

}
