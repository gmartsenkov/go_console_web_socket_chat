package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

//Connections ...
type Connections struct {
	Connections []*websocket.Conn
	Messages    []string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var connections = Connections{}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err.Error())
	}
	go func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			fmt.Println(connections.Connections)
			for _, ws := range connections.Connections {
				fmt.Println(string(msg))
				fmt.Println(ws)
				connections.Messages = append(connections.Messages, string(msg))
				ws.WriteMessage(1, msg)
			}
		}
	}(conn)

}

func connectHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Println("connected", conn)
	connections.Connections = append(connections.Connections, conn)
	sendOldMessages(conn)
}

func sendOldMessages(ws *websocket.Conn) {
	for i, msg := range connections.Messages {
		fmt.Println("------------------message-------------------")
		fmt.Println(i, msg)
		ws.WriteMessage(1, []byte(msg))
	}
}

func main() {
	http.HandleFunc("/echo", echoHandler)
	http.HandleFunc("/connect", connectHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}
