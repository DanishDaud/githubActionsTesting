package service

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"gopkg.in/mgo.v2/bson"
)

type AddBalanceMessage struct {
	UserID bson.ObjectId `json:"userid,omitempty" bson:"userid"`
	Amount int64         `json:"amount,omitempty" bson:"amount"`
}

type DeductBalanceMessage struct {
	UserID bson.ObjectId `json:"userid,omitempty" bson:"userid"`
	Amount float32       `json:"amount,omitempty" bson:"amount"`
}

type BillingManager struct {
	SVC *sqs.SQS
}

func NewBillingManager() *BillingManager {
	// Setting up SQS svc
	region := os.Getenv("AWS_REGION")
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := sqs.New(sess)
	bm := &BillingManager{SVC: svc}
	return bm
}

func (bm *BillingManager) ProcessAddBalance(userId bson.ObjectId, amount int64) {
	balanceObj := AddBalanceMessage{Amount: amount, UserID: userId}

	b, err := json.Marshal(balanceObj)
	if err != nil {
		return
	}
	//now := time.Now().Add(1 * time.Minute).UTC()
	//s := rand.NewSource(now.UnixNano())
	//rnd := rand.New(s)
	//rn := rnd.Int63n(8) + 12

	params := &sqs.SendMessageInput{
		MessageBody:            aws.String(string(b)),
		QueueUrl:               aws.String(os.Getenv("BALANCE_ADD_QUEUE")),
		MessageDeduplicationId: aws.String(bson.NewObjectId().Hex()),
		MessageGroupId:         aws.String(userId.Hex()),
		//DelaySeconds: aws.Int64(rn),
	}
	if _, err := bm.SVC.SendMessage(params); err != nil {
		logrus.Errorln("Error : ", err.Error())
		return
	}
}

func (bm *BillingManager) ProcessDeductBalance(userId string, amount float32) error {
	//if campaignId == "TEST" {
	//	return nil
	//}

	//if !bson.IsObjectIdHex(userId) {
	//	return errors.New("user id is invalid")
	//}
	//
	//if !bson.IsObjectIdHex(campaignId) {
	//	return errors.New("campaign id is invalid")
	//}
	//

	//var campaignIdHex = bson.ObjectIdHex(campaignId)
	var userIdHex bson.ObjectId = bson.ObjectIdHex(userId)

	balanceObj := DeductBalanceMessage{Amount: amount, UserID: userIdHex}

	b, err := json.Marshal(balanceObj)
	if err != nil {
		return err
	}
	//now := time.Now().Add(1 * time.Minute).UTC()
	//s := rand.NewSource(now.UnixNano())
	//rnd := rand.New(s)
	//rn := rnd.Int63n(8) + 12

	params := &sqs.SendMessageInput{
		MessageBody:            aws.String(string(b)),
		QueueUrl:               aws.String(BalanceDeductQueue()),
		MessageDeduplicationId: aws.String(bson.NewObjectId().Hex()),
		MessageGroupId:         aws.String(userId),
		//DelaySeconds: aws.Int64(rn),
	}
	if _, err := bm.SVC.SendMessage(params); err != nil {
		logrus.Errorln("Error : ", err.Error())
		return err
	}

	return nil
}

func BalanceDeductQueue() string {
	return os.Getenv("BALANCE_DEDUCT_QUEUE")
}
