package controller

import (
	"net/http"
)

var ws *websocketController

// Initialize Start the web apps up
func Initialize(url string) {

	ws = newWebsocketController(url)

	registerRoutes()

	registerFileServers()
}

func registerRoutes() {
	http.HandleFunc("/ws", ws.handleMessage)
}

func registerFileServers() {
	http.Handle("/public/",
		http.FileServer(http.Dir("assets")))
	http.Handle("/public/lib/",
		http.StripPrefix("/public/lib/",
			http.FileServer(http.Dir("node_modules"))))
}
