package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jroimartin/gocui"
)

const MESSAGE_ENDPOINT = "ws://localhost:8080/message"
const SUBSCRIBE_ENDPOINT = "ws://localhost:8080/subscribe"

var (
	USERNAME = ""
)

func setup() {
	fmt.Print("Enter your name: ")
	reader := bufio.NewReader(os.Stdin)
	USERNAME, _ = reader.ReadString('\n')
	USERNAME = strings.Replace(USERNAME, "\n", "", -1)
}

func setKeyBindings(g *gocui.Gui) {
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

func main() {
	setup()
	g := gocui.NewGui()
	if err := g.Init(); err != nil {
		log.Panicln(err)
	}
	g.Cursor = true
	defer g.Close()

	g.SetLayout(layout)
	setKeyBindings(g)
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
	chatView, err := g.View("chat")
	if err != nil {
		return err
	}
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(MESSAGE_ENDPOINT, nil)
	message := USERNAME + ": (" + time.Now().Format(time.Kitchen) + ") : " + v.Buffer()
	if err != nil {
		fmt.Fprint(chatView, "Server is not responding")
		return
	}
	ws.WriteMessage(1, []byte(message))
	v.Clear()
	v.SetCursor(0, 0)
	return nil
}

func listenForMessages(gui *gocui.Gui) {
	v, _ := gui.View("chat")
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(SUBSCRIBE_ENDPOINT, nil)
	if err != nil {
		fmt.Fprint(v, "Server is not responding")
	}
	ws.WriteMessage(1, []byte(USERNAME))
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

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
}
