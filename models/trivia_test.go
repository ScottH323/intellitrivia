package models

import (
	"testing"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestQuery(t *testing.T) {
	q := Question{
		Text: "“ Baby ” , “ Posh ” and “ Scary ” are nicknames of members from which music act ? ",
		Options: []*Option{
			&Option{Text: "Space Gulls"},
			&Option{Text: "Seasoning Babies"},
			&Option{Text: "Spice Girls"},
		},
	}

	answer := q.Solve()

	logrus.WithFields(logrus.Fields{
		"answer":      answer.Text,
		"count":       answer.ResultCount,
		"probability": q.Probability(answer),
	}).Info("Correct Answer")

	if answer.Text != "Spice Girls" {
		t.Fail()
	}
}
