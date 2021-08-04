package service

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

var (
	svc *sqs.SQS
	log *logrus.Logger
)

func init() {
	// Setting up SQS svc
	sess, err := session.NewSession()
	if err != nil {
		log.Fatal("Failed to create SQS session", err)
		return
	}
	svc = sqs.New(sess)
}

type BalanceMessage struct {
	UserID     bson.ObjectId `json:"userid,omitempty" bson:"userid"`
	Amount     float32       `json:"amount,omitempty" bson:"amount"`
	ExpiryDate time.Time     `json:"expiry,omitempty" bson:"expiry"`
}

func AddBalance(msg *BalanceMessage) error {
	if msg == nil {
		return errors.New("Message is empty")
	}

	message, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	params := &sqs.SendMessageInput{
		MessageBody: aws.String(string(message)),
		QueueUrl:    aws.String(os.Getenv("https://sqs.us-east-2.amazonaws.com/092914350126/AddBalanceQueue.fifo")),
	}
	resp, err := svc.SendMessage(params)
	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		logrus.Errorln("Error while sending SQS message", err)
		return err
	}

	logrus.WithField("sqsResp", resp).Infoln("message sent")

	return nil
}
