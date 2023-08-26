package main

import (
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/django/v3"
)

var channels = make([]*websocket.Conn, 0)
var register = make(chan *websocket.Conn)
var unregister = make(chan *websocket.Conn)

func WebsocketHandler(c *websocket.Conn) {
	defer func() {
		log.Printf("closed Connection from %s\n", c.RemoteAddr().String())
		unregister <- c
		c.Close()
	}()

	log.Printf("new Websocket connection from %s\n", c.RemoteAddr().String())
	register <- c

	for {
		_, message, err := c.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
			}

			return
		}

		log.Printf("received : %s\n", message)
	}
}

func SocketListen() {
	go func() {
		for {
			select {
			case c := <-register:
				channels = append(channels, c)
			case c := <-unregister:
				for i, con := range channels {
					if con == c {
						channels = append(channels[:i], channels[i+1:]...)
					}
				}
			}
		}
	}()

	for {
		for _, c := range channels {
			// err := c.WriteMessage(websocket.TextMessage, []byte("hello"))

			// if err != nil {
			// 	log.Println(err)
			// 	unregister <- c
			// }
		}

		time.Sleep(time.Second)
	}
}

func main() {
	templateEngine := django.New("./views", ".html")

	go SocketListen()

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

	app.Get("/ws", websocket.New(WebsocketHandler))

	err := app.Listen(":3000")

	log.Fatal(err)
}
