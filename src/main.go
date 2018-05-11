
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/melonmanchan/asd/src/google"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func checkOrigin(r *http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{
	CheckOrigin: checkOrigin,
} // use default options

func readFile(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		r := bytes.NewReader(message)

		labels, err := google.ReaderToFaceResults(r)

		if err != nil {VJ
			log.Printf("Failed to detect labels: %v", err)
		} else {
			fmt.Println("Labels:")

			for _, label := range labels {
				c.WriteJSON(label)
				fmt.Printf("Confidence: %f\n", label.DetectionConfidence)
				fmt.Printf("Anger: %s\n", label.AngerLikelihood)
				fmt.Printf("Blurred: %s\n", label.BlurredLikelihood)
				fmt.Printf("Joy: %s\n", label.JoyLikelihood)
				fmt.Printf("Sorrow: %s\n", label.SorrowLikelihood)
				fmt.Printf("Surprise: %s", label.SurpriseLikelihood)
			}
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", readFile)
	log.Printf("Listening at %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
