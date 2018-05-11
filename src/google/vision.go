package google

import (
	"context"
	"io"

	vision "cloud.google.com/go/vision/apiv1"
	proto "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

// ReaderToFaceResults ...
func ReaderToFaceResults(reader io.Reader) ([]*proto.FaceAnnotation, error) {
	ctx := context.Background()
	// Creates a client.
	client, err := vision.NewImageAnnotatorClient(ctx)

	if err != nil {
		return nil, err
	}

	image, err := vision.NewImageFromReader(reader)
	if err != nil {
		return nil, err
	}

	labels, err := client.DetectFaces(ctx, image, nil, 1)

	if err != nil {
		return nil, err
	}

	return labels, nil
}
