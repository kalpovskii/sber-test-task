package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
)

func initConfig() {
	viper.SetEnvPrefix("CHECKLIST")
	viper.AutomaticEnv()
}

func main() {
	initConfig()

	broker := viper.GetString("KAFKA_BROKER")
	topic := viper.GetString("KAFKA_TOPIC")
	logFile := viper.GetString("KAFKA_LOG_FILE")

	if broker == "" || topic == "" || logFile == "" {
		log.Fatal("KAFKA_BROKER, KAFKA_TOPIC or KAFKA_LOG_FILE is not configured")
	}

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer file.Close()

	logger := log.New(file, "", log.LstdFlags)
	logger.Println("Kafka Logger started")

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: "kafka-logger-group",
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			logger.Printf("error reading message: %v\n", err)
			continue
		}

		logger.Printf("[%s] %s\n", time.Now().Format(time.RFC3339), string(m.Value))
	}
}