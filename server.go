package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type User struct {
	name       string `json:"name"`
	connection *websocket.Conn
}

type Channel struct {
	Users          []*User
	MessageHistory [][]byte
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var MainChannel = Channel{}

func main() {
	http.HandleFunc("/message", messageHandler)
	http.HandleFunc("/subscribe", subscribeHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}

func messageHandler(w http.ResponseWriter, r *http.Request) {
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
			for _, user := range MainChannel.Users {
				fmt.Println(string(msg))
				MainChannel.MessageHistory = append(MainChannel.MessageHistory, msg)
				user.connection.WriteMessage(1, msg)
			}
		}
	}(conn)

}

func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Print(err.Error())
	}
	user := User{}
	conn.ReadJSON(&user)
	fmt.Println(user)
	MainChannel.Users = append(MainChannel.Users, &User{name: user.name, connection: conn})
	go notifyChannel(user.name)
	go sendOldMessages(conn)
}

func notifyChannel(userName string) {
	for _, user := range MainChannel.Users {
		user.connection.WriteMessage(1, []byte(userName+" has connected\n"))
	}
}

func sendOldMessages(ws *websocket.Conn) {
	for _, msg := range MainChannel.MessageHistory {
		time.Sleep(100 * time.Millisecond)
		ws.WriteMessage(1, msg)
	}
}
