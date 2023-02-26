package main

import (
	"github.com/gofiber/fiber/v2"
	ws "github.com/gofiber/websocket/v2"
)

func main() {

	hub := CreateHub()

	app := fiber.New()

	app.Static("/", "./public")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./render/index.html")
	})

	app.Get("/:session", func(c *fiber.Ctx) error {
		return c.SendFile("./render/session.html")
	})

	app.Get("/ws/:session",
		func(c *fiber.Ctx) error { // check if a client can establish websocket connection
			if ws.IsWebSocketUpgrade(c) {
				c.Next()
			}
			return nil
		},
		ws.New(func(conn *ws.Conn) {
			WebsocketConnectionLoop(hub, conn)
		}))

	app.Listen(":3005")
}
