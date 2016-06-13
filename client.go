package main

import (
	"bufio"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

import "github.com/jroimartin/gocui"

import "os"

import "strings"

//import "strings"

var (
	origin      = "http://localhost/"
	url         = "ws://localhost:8080/message"
	connect_url = "ws://localhost:8080/subscribe"
	ctr         = 0
	_, _        = fmt.Print("What is your name: ")
	reader      = bufio.NewReader(os.Stdin)
	NAME, _     = reader.ReadString('\n')
)

func main() {
	g := gocui.NewGui()
	if err := g.Init(); err != nil {
		log.Panicln(err)
	}
	g.Cursor = true
	defer g.Close()

	g.SetLayout(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, inputReader); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v_chat, err := g.SetView("chat", 0, 0, maxX-1, (maxY/2)+(maxY/3)); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v_chat.Autoscroll = true
		go listenForMessages(g)
		//go listenChatFile(v_chat)
	}

	if v, err := g.SetView("input", 0, (maxY/2)+maxY/3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = true
		v.Title = "Say something.."
	}
	g.SetCurrentView("input")

	return nil
}

func inputReader(g *gocui.Gui, v *gocui.View) (err error) {
	chat_view, err := g.View("chat")
	if err != nil {
		return err
	}
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(url, nil)
	name := strings.Replace(NAME, "\n", "", -1)
	message := name + ": (" + time.Now().Format(time.Kitchen) + ") : " + v.Buffer()
	if err != nil {
		fmt.Fprint(chat_view, "Server is not responding")
		return
	}
	user := User{name: name, message: message}
	fmt.Println(user)
	ws.WriteJSON(user)
	v.Clear()
	v.SetCursor(0, 0)
	return nil
}

func listenForMessages(gui *gocui.Gui) {
	v, _ := gui.View("chat")
	fmt.Print(websocket.BinaryMessage)
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(connect_url, nil)
	if err != nil {
		fmt.Fprint(v, "Server is not responding")
	}
	for {
		_, msg, err := ws.ReadMessage()
		if err == nil {
			gui.Execute(func(gui *gocui.Gui) error {
				v, _ := gui.View("chat")
				fmt.Fprint(v, string(msg))
				return nil
			})
		}
	}
}

type User struct {
	name    string `json:"name"`
	message string `json:"message"`
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
}
