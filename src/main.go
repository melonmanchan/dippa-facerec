// Sample vision-quickstart uses the Google Cloud Vision API to label an image.
package main

import (
	"fmt"
	"log"
	"os"

	"./google"
)

func main() {
	// Sets the name of the image file to annotate.
	filename := "./smile.jpg"

	file, err := os.Open(filename)

	defer file.Close()

	labels, err := google.ReaderToFaceResults(file)
	if err != nil {
		log.Fatalf("Failed to detect labels: %v", err)
	}

	fmt.Println("Labels:")
	for _, label := range labels {
		fmt.Printf("Confidence: %f\n", label.DetectionConfidence)
		fmt.Printf("Anger: %s\n", label.AngerLikelihood)
		fmt.Printf("Blurred: %s\n", label.BlurredLikelihood)
		fmt.Printf("Joy: %s\n", label.JoyLikelihood)
		fmt.Printf("Sorrow: %s\n", label.SorrowLikelihood)
		fmt.Printf("Surprise: %s", label.SurpriseLikelihood)
	}
}
