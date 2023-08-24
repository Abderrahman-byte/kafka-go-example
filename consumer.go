package main

import (
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func main() {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:19092",
		"group.id":          "kafka-go-consumer",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Consumer created")

	c.SubscribeTopics([]string{"logs"}, nil)

	log.Println("Consumer subscribed")

	for true {
		msg, err := c.ReadMessage(time.Second)

		if err != nil && !err.(kafka.Error).IsTimeout() {
			log.Printf("Consumer error: %v (%v)\n", err, msg)
			continue
		} else if err != nil {
			continue
		}

		log.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
	}

	c.Close()
}
