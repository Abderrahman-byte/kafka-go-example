package main

import (
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/django/v3"
)

func main() {
	templateEngine := django.New("./views", ".html")

	app := fiber.New(fiber.Config{
		ServerHeader:      "Abderrahmane",
		AppName:           "Kafka go example v0.0.1",
		CaseSensitive:     true,
		Views:             templateEngine,
		PassLocalsToViews: true,
	})

	app.Static("/static", "./static/")

	app.Use(logger.New())
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("WS-allowed", true)
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Get("", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"title": "Kafka example",
		}, "layouts/base")
	})

	err := app.Listen(":3000")

	log.Fatal(err)
}


func WebsocketLogs () {

}
