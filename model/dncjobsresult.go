package model

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type DNCJobsResultList []DNCJobResult
type DNCJobResult struct {
	ID            bson.ObjectId  `json:"id" bson:"_id"`
	DncJobID      bson.ObjectId  `json:"dncJobId" bson:"dncJobId"`
	ContactListId *bson.ObjectId `json:"contactListId" bson:"contactListId"`
	TTSListID     *bson.ObjectId `json:"ttsListID" bson:"ttsListID"`
	Matched       []string       `json:"matched" bson:"matched"`
	Clean         []string       `json:"clean" bson:"clean"`
	JobId         string         `json:"jobid" bson:"jobid"`
	CreateDate    time.Time      `json:"createDate" bson:"createDate"`
	UpdateDate    time.Time      `json:"updateDate" bson:"updateDate"`
}
