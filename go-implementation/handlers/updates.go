package handlers

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true }, // TODO remove for prod
}

func (app *AppContext) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for measurement := range app.Updates.Subscribe() {
		if err := conn.WriteJSON(measurement); err != nil {
			log.Println("write error:", err)
			break
		}
	}
}
