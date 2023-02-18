package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	ws "github.com/gofiber/websocket/v2"
)

type channelSet struct {
	register   chan *ws.Conn
	unregister chan *ws.Conn
	broadcast  chan string
}
type client struct{}

var clients = make(map[*ws.Conn]client)

func sendBroadcastMessage(cs *channelSet, message string) {
	for c := range clients {
		if err := c.WriteMessage(ws.TextMessage, []byte(message)); err != nil {
			fmt.Println("write error: err")
			c.WriteMessage(ws.CloseMessage, []byte{})
			c.Close()
		}
	}
}

func userHub(cs *channelSet) {
	for {
		select {
		case registerUser := <-cs.register:
			clients[registerUser] = client{}
			fmt.Println("new client is connected")

		case unregisterUser := <-cs.unregister:
			delete(clients, unregisterUser)
			fmt.Println("connection terminated")
		}
	}
}

func messageHub(cs *channelSet) {
	for {
		select {
		case message := <-cs.broadcast:
			fmt.Println("message received")
			sendBroadcastMessage(cs, message)
		}
	}
}

func main() {
	var channels = channelSet{
		register:   make(chan *ws.Conn),
		unregister: make(chan *ws.Conn),
		broadcast:  make(chan string),
	}

	go userHub(&channels)
	go messageHub(&channels)

	app := fiber.New()

	app.Static("/", "./public")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./render/index.html")
	})

	app.Get("/ws/",
		func(c *fiber.Ctx) error {
			if ws.IsWebSocketUpgrade(c) {
				c.Next()
			}
			return nil
		},
		ws.New(func(c *ws.Conn) {
			defer func() {
				channels.unregister <- c
				c.Close()
			}()

			channels.register <- c

			for {
				messageType, message, err := c.ReadMessage()
				if err != nil {
					if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
						fmt.Println("read error:", err)
					}
					return
				}

				if messageType == ws.TextMessage {
					channels.broadcast <- string(message)
				} else {
					fmt.Println("websocket message recived of type", messageType)
				}
			}
		}))

	app.Listen(":3000")
}
