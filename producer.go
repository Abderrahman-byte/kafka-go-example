package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type LogRecord struct {
	id        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

func main() {
	topic := "logs"
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:29092"})

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	for i := 1; i <= 100; i++ {
		message := &LogRecord{
			id:        i,
			Timestamp: time.Now(),
			Level:     "INFO",
			Message:   "hello",
		}

		b, err := json.Marshal(message)

		if err != nil {
			log.Println(err)
			continue
		}

		err = p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          b,
		}, nil)

		if err != nil {
			log.Fatal(err)
		}

		p.Flush(15 * 1000)
		time.Sleep(time.Second)
	}

}
