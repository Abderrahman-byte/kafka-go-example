#!/bin/bash

# Wait for Kafka to become available
sleep 20

# Create the Kafka topic
kafka-topics --create --if-not-exists --topic chat --bootstrap-server kafka-1:9092 --partitions 1 --replication-factor 2

echo "Kafka topic 'chat' created."
