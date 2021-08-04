package model

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type (
	// Contact list represents the structure of our resource
	SoundFileInfo struct {
		ID               bson.ObjectId `json:"id" bson:"_id"`
		Name             string        `json:"name" bson:"name"`
		SoundFileCode    int64         `json:"soundFileCode" bson:"soundFileCode"`
		FileName         string        `json:"fileName" bson:"fileName"`
		FileSize         string        `json:"fileSize" bson:"fileSize"`
		FileType         string        `json:"fileType" bson:"fileType"`
		OriginalFilePath string        `json:"originalFilePath" bson:"originalFilePath"`
		WavFilePath      string        `json:"wavFilePath" bson:"wavFilePath"`
		PCMAFilePath     string        `json:"pcmaFilePath" bson:"pcmaFilePath"`
		Duration         int           `json:"duration" bson:"duration"`
		UserID           bson.ObjectId `json:"userid" bson:"userid"`
		CreateDate       time.Time     `json:"createDate" bson:"createDate"`
		UpdateDate       time.Time     `json:"updateDate" bson:"updateDate"`
	}
)

type SoundFileInfos []SoundFileInfo
