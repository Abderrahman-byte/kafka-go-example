package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/django/v3"
	"github.com/google/uuid"
)

type Message struct {
	Type    string `json:"type"`
	From    string `json:"from"`
	Content string `json:"content"`
}

type Config struct {
	Addr        string
	KafkaBroker string
}

var connections = make([]*websocket.Conn, 0)
var register = make(chan *websocket.Conn)
var unregister = make(chan *websocket.Conn)
var messages = make(chan Message)
var produce = make(chan []byte, 3)

func SendMessage(message Message) {
	data, err := json.Marshal(message)

	if err != nil {
		log.Println("SendMessage-ERROR:", err)
		return
	}

	produce <- data
}

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

	SendMessage(startNotification)

	for {
		_, message, err := c.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("READ-ERROR:", err)
			}

			return
		}

		SendMessage(Message{
			Type:    "message",
			From:    fmt.Sprintf("User-%d", userId),
			Content: string(message),
		})
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

func KafkaProduct(broker string) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})
	CheckError(err)

	topic := "chat"

	for {
		select {
		case data := <-produce:
			go func() {
				err = p.Produce(&kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
					Value:          data,
				}, nil)

				if err != nil {
					log.Println("PUBLISHER ERROR", err)
				}

				p.Flush(15 * 1000)
			}()
		}

	}
}

func KafkaConsume(broker string) {
	id := uuid.New()

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        broker,
		"auto.offset.reset":        "latest",
		"allow.auto.create.topics": true,
		"group.id":                 id.String(),
	})

	CheckError(err)

	c.SubscribeTopics([]string{"chat"}, nil)

	for {
		msg, err := c.ReadMessage(time.Second)

		if err != nil && !err.(kafka.Error).IsTimeout() {
			log.Printf("CONSUMER-ERROR: %v (%v)\n", err, msg)
			continue
		} else if err != nil {
			continue
		}

		message := Message{}

		err = json.Unmarshal(msg.Value, &message)

		if err != nil {
			log.Printf("UNMARSHAL-ERROR: %v \n", err)
			continue
		}

		messages <- message
	}
}

func main() {
    config := InitConfig()
	app := InitApp()

	go SocketListen()
	go KafkaConsume(config.KafkaBroker)
	go KafkaProduct(config.KafkaBroker)

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

	log.Fatal(app.Listen(config.Addr))
}

func InitConfig() Config {
    address := os.Getenv("ADDRESS")
    kafka := os.Getenv("KAFKA")

    if address == "" {
        address = ":3000"
    }

    if kafka == "" {
        kafka = "localhost:19092"
    }

    return Config{address, kafka}
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
