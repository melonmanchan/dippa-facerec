package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

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

	var conn *amqp.Connection

	for {
		newConn, err := amqp.Dial(rabbitConn)

		if err == nil {
			conn = newConn
			break
		}

		log.Printf("Connection failed :%s\n", err)
		time.Sleep(3 * time.Second)
	}

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

	readQueue, err := ch.QueueDeclare(
		"",    // name
		true,  // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	writeQueue, err := ch.QueueDeclare(
		"google_results", // name
		true,             // durable
		false,            // delete when usused
		true,             // exclusive
		false,            // no-wait
		nil,              // arguments
	)

	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		readQueue.Name, // queue name
		"",             // routing key
		"images",       // exchange
		false,
		nil)

	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		readQueue.Name, // queue
		"",             // consumer
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)

	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		log.Println("Listening for messages...")
		for d := range msgs {
			processingData := &types.ProcessingData{}
			googleFacialRecognitionRes := &types.GoogleFacialRecognition{}

			if err := proto.Unmarshal(d.Body, processingData); err != nil {
				log.Print("Failed to parse input: ", err)
				break
			}

			googleFacialRecognitionRes.User = processingData.User
			googleFacialRecognitionRes.Image = processingData.Contents

			labels, err := google.ReaderToFaceResults(bytes.NewReader(processingData.Contents))

			if err != nil {
				log.Printf("Failed to detect labels: %v", err)
			} else {
				fmt.Println("Labels:")

				for _, label := range labels {

					emotion := &types.GoogleEmotion{}

					fmt.Printf("Confidence: %f\n", label.DetectionConfidence)
					fmt.Printf("Anger: %s\n", label.AngerLikelihood)
					fmt.Printf("Blurred: %s\n", label.BlurredLikelihood)
					fmt.Printf("Joy: %s\n", label.JoyLikelihood)
					fmt.Printf("Sorrow: %s\n", label.SorrowLikelihood)
					fmt.Printf("Surprise: %s\n", label.SurpriseLikelihood)

					emotion.DetectionConfidence = label.DetectionConfidence

					emotion.Anger = float32(label.AngerLikelihood)
					emotion.Blurred = float32(label.BlurredLikelihood)
					emotion.Joy = float32(label.JoyLikelihood)
					emotion.Sorrow = float32(label.SorrowLikelihood)
					emotion.Surprise = float32(label.SurpriseLikelihood)

					googleFacialRecognitionRes.Emotion = emotion
				}

				out, err := proto.Marshal(googleFacialRecognitionRes)

				if err != nil {
					log.Println("processing error", err)
					break
				}

				err = ch.Publish(
					"google_results", // exchange
					writeQueue.Name,  // routing key
					false,            // mandatory
					false,            // immediate
					amqp.Publishing{
						Body: out,
					})

				if err != nil {
					log.Println("writing results error:", err)
					break
				}
			}
		}
	}()

	<-forever
}
