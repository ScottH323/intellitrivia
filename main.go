package main

import (
	"github.com/sirupsen/logrus"
	"cloud.google.com/go/vision/apiv1"
	"os"
	"context"
	"strings"
	"itellitrivia/models"
)

func init() {
	logrus.Info("Init")
}

func main() {
	ctx := context.Background()

	// Creates a client.
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		logrus.Fatalf("Failed to create client: %v", err)
	}

	// Sets the name of the image file to annotate.
	filename := "./testdata/space.jpg" //TODO

	file, err := os.Open(filename)
	if err != nil {
		logrus.Fatalf("Failed to read file: %v", err)
	}

	image, err := vision.NewImageFromReader(file)
	if err != nil {
		logrus.Fatalf("Failed to create image: %v", err)
	}

	doc, err := client.DetectDocumentText(ctx, image, nil)
	if err != nil {
		logrus.WithError(err).Error("Unable to get document")
	}

	blocks := make(map[int]string)
	for _, p := range doc.Pages {
		for k, b := range p.Blocks {
			logrus.WithField("block_number", k).Debug("Checking Block")
			tBlock := []string{}

			for _, pa := range b.Paragraphs {
				for _, w := range pa.Words {
					tWord := []string{}

					for _, s := range w.Symbols {
						tWord = append(tWord, s.Text)
					}

					tBlock = append(tBlock, strings.Join(tWord, ""))
				}
			}

			blocks[k] = strings.Join(tBlock, " ")
			logrus.WithField("block", blocks[k]).Debug("Block Text")
		}
	}

	var questionKey int

	for k, b := range blocks {
		if strings.Contains(b, "?") {
			questionKey = k
			break
		}
	}

	q := models.Question{
		Text: blocks[questionKey],
		Options: []*models.Option{
			&models.Option{Text: blocks[questionKey+1]},
			&models.Option{Text: blocks[questionKey+2]},
			&models.Option{Text: blocks[questionKey+3]},
		},
	}

	logrus.Printf("Q: %s", q.Text)
	for k, o := range q.Options {
		logrus.Printf("A%v: %s", k, o.Text)
	}

	answer := q.Solve() //See if this works :)

	logrus.WithFields(logrus.Fields{
		"answer":      answer.Text,
		"count":       answer.ResultCount,
		"probability": q.Probability(answer),
	}).Info("Correct Answer")
}
