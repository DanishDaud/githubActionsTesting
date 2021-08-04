package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type DNCJobType int8

const (
	DNCScrubJobTypeProcessing DNCJobType = iota
	DNCScrubJobTypeCompleted             = 1
)

type DNCJobsList []DNCJobs

type DNCJobs struct {
	ID            bson.ObjectId  `json:"id" bson:"_id"`
	Status        DNCJobType     `json:"status" bson:"status"`
	ContactListId *bson.ObjectId `json:"contactListId,omitempty" bson:"contactListId,omitempty"`
	TTSListID     *bson.ObjectId `json:"ttsListID,omitempty" bson:"ttsListID,omitempty"`
	JobId         string         `json:"jobid" bson:"jobid"`
	CreateDate    time.Time      `json:"createDate" bson:"createDate"`
	UpdateDate    time.Time      `json:"updateDate" bson:"updateDate"`
}
