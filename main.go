package main

import (
	"github.com/gofiber/fiber/v2"
	ws "github.com/gofiber/websocket/v2"
)

func main() {

	clientDataStore := MakeClientDataStore()
	hub := CreateHub(clientDataStore)

	app := fiber.New()

	app.Static("/", "./public")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./render/index.html")
	})

	app.Get("/:session", func(c *fiber.Ctx) error {
		return c.SendFile("./render/session.html")
	})

	app.Get("/ws/:session",
		func(c *fiber.Ctx) error {
			if ws.IsWebSocketUpgrade(c) {
				c.Next()
			}
			return nil
		},
		ws.New(func(conn *ws.Conn) {
			WebsocketConnectionLoop(clientDataStore, hub, conn)
		}))

	app.Listen(":3005")
}
