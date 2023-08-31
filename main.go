package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/django/v3"
)

type Message struct {
	Type    string `json:"type"`
	From    string `json:"from"`
	Content string `json:"content"`
}

var connections = make([]*websocket.Conn, 0)
var register = make(chan *websocket.Conn)
var unregister = make(chan *websocket.Conn)
var messages = make(chan Message)

func WebsocketHandler(c *websocket.Conn) {
	userId := rand.Intn(100)

	defer func() {
		log.Printf("closed Connection from %s\n", c.RemoteAddr().String())
		unregister <- c
		c.Close()
	}()

	log.Printf("new Websocket connection from %s\n", c.RemoteAddr().String())
	register <- c

	startNotification := Message{
		Type:    "notification",
		From:    "App",
		Content: fmt.Sprintf("User-%d has joined the chat", userId),
	}

    messages <- startNotification

	for {
		_, message, err := c.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("READ-ERROR:", err)
			}

			return
		}

		messages <- Message{
			Type:    "message",
			From:    fmt.Sprintf("User-%d", userId),
			Content: string(message),
		}
	}
}

func SocketListen() {
	for {
		select {
		case c := <-register:
			connections = append(connections, c)
		case c := <-unregister:
			for i, con := range connections {
				if con == c {
					connections = append(connections[:i], connections[i+1:]...)
				}
			}
		case msg := <-messages:
			data, err := json.Marshal(msg)

			if err != nil {
				log.Println("ERROR-JSON:", err)
				continue
			}

			for _, c := range connections {
				err := c.WriteMessage(websocket.TextMessage, data)

				if err != nil {
					log.Println(err)
					unregister <- c
				}
			}
		}
	}
}

func main() {
	app := InitApp()

	go SocketListen()

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("WS-ALLOWED", true)
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Get("", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"title": "Chat app",
		})
	})

	app.Get("/ws", websocket.New(WebsocketHandler))

	log.Fatal(app.Listen(":3000"))
}

func InitApp() *fiber.App {
	templateEngine := django.New("./views", ".html")
	templateEngine.Reload(true)

	app := fiber.New(fiber.Config{
		ServerHeader:      "Abderrahmane",
		AppName:           "Kafka go example v0.0.1",
		CaseSensitive:     true,
		Views:             templateEngine,
		PassLocalsToViews: true,
		ViewsLayout:       "layouts/base",
	})

	app.Static("/static", "./static/")

	app.Use(logger.New())

	return app
}

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
