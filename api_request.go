package main

import "golang.org/x/net/websocket"

type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc RequestProcessor
}

type wsRoute struct {
	Path        string
	HandlerFunc func(*websocket.Conn)
}

func (h *httpApiHandler) getRoutes() []route {
	return []route{
		{
			"wake",
			"POST",
			"/wake",
			h.Wake,
		},
		{
			"halt",
			"POST",
			"/halt",
			h.Halt,
		},
		{
			"status",
			"GET",
			"/status/{host}",
			h.Status,
		},
	}
}

func (h *httpApiHandler) getWSRoutes() []wsRoute {
	return []wsRoute{
		{
			"/status-stream",
			h.StatusStream,
		},
	}
}
