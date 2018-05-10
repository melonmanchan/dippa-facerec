// Sample vision-quickstart uses the Google Cloud Vision API to label an image.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
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
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/ws", readFile)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

//	filename := "./smile.jpg"
//
//	file, err := os.Open(filename)
//
//	defer file.Close()
//
//	labels, err := google.ReaderToFaceResults(file)
//	if err != nil {
//		log.Fatalf("Failed to detect labels: %v", err)
//	}
//
//	fmt.Println("Labels:")
//	for _, label := range labels {
//		fmt.Printf("Confidence: %f\n", label.DetectionConfidence)
//		fmt.Printf("Anger: %s\n", label.AngerLikelihood)
//		fmt.Printf("Blurred: %s\n", label.BlurredLikelihood)
//		fmt.Printf("Joy: %s\n", label.JoyLikelihood)
//		fmt.Printf("Sorrow: %s\n", label.SorrowLikelihood)
//		fmt.Printf("Surprise: %s", label.SurpriseLikelihood)
//	}
