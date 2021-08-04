package model

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	DSoundFileDNC      = 1
	DSoundFileTransfer = 2
	DSoundFileOptOut   = 3
)

type (
	// Contact list represents the structure of our resource
	DefaultSoundFileInfo struct {
		ID               bson.ObjectId `json:"id" bson:"_id"`
		Name             string        `json:"name" bson:"name"`
		Type             int           `json:"type" bson:"type"`
		FileName         string        `json:"fileName" bson:"fileName"`
		OriginalFilePath string        `json:"originalFilePath" bson:"originalFilePath"`
		WavFilePath      string        `json:"wavFilePath" bson:"wavFilePath"`
		PCMAFilePath     string        `json:"pcmaFilePath" bson:"pcmaFilePath"`
		CreateDate       time.Time     `json:"createDate" bson:"createDate"`
		UpdateDate       time.Time     `json:"updateDate" bson:"updateDate"`
	}
)

type DefaultSoundFileInfos []DefaultSoundFileInfo
