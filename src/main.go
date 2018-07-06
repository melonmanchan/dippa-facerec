package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	proto "github.com/golang/protobuf/proto"
	google "github.com/melonmanchan/dippa-facerec/src/google"
	types "github.com/melonmanchan/dippa-proto/build/go"
	"github.com/streadway/amqp"
)

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)

	if !exists {
		value = fallback
	}

	return value
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func main() {
	var rabbitConn = getEnv("RABBITMQ_ADDRESS", "amqp://guest:guest@localhost:5672/")
	conn, err := amqp.Dial(rabbitConn)

	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"images", // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)

	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		true,  // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,   // queue name
		"",       // routing key
		"images", // exchange
		false,
		nil)

	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			processingData := &types.ProcessingData{}

			if err := proto.Unmarshal(d.Body, processingData); err != nil {
				log.Print("Failed to parse address book:", err)
				break
			}

			labels, err := google.ReaderToFaceResults(bytes.NewReader(processingData.Contents))

			if err != nil {
				log.Printf("Failed to detect labels: %v", err)
			} else {
				fmt.Println("Labels:")

				for _, label := range labels {
					fmt.Printf("Confidence: %f\n", label.DetectionConfidence)
					fmt.Printf("Anger: %s\n", label.AngerLikelihood)
					fmt.Printf("Blurred: %s\n", label.BlurredLikelihood)
					fmt.Printf("Joy: %s\n", label.JoyLikelihood)
					fmt.Printf("Sorrow: %s\n", label.SorrowLikelihood)
					fmt.Printf("Surprise: %s\n", label.SurpriseLikelihood)
				}
			}
		}
	}()

	<-forever
}
