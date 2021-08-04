package model

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type (
	ContactListStatus     int8
	ContactListType       int8
	DNCScrubProcessStatus int8
)

const (
	ContactListStatusNew        ContactListStatus = 1
	ContactListStatusProcessing                   = 2
	ContactListStatusActive                       = 3
	ContactListStatusError                        = 4
)

const (
	ContactListTypeDefault      ContactListType = 1
	ContactListTypeCallerGroups                 = 2
	ContactListTypeDNC                          = 3
)

const (
	DNCScrubProcessStatusStarted DNCScrubProcessStatus = iota
	DNCScrubProcessStatusAllBatchesSent
	DNCScrubProcessStatusResponseReceived
)

type (
	ContactList struct {
		ID             bson.ObjectId         `json:"id" bson:"_id"`
		UserID         bson.ObjectId         `json:"userid,omitempty" bson:"userid"`
		Status         ContactListStatus     `json:"status" bson:"status"`
		DNCScrubStatus DNCScrubProcessStatus `json:"dncScrubStatus" bson:"dncScrubStatus"`
		Type           ContactListType       `json:"type" bson:"type"`
		Imported       int64                 `json:"imported" bson:"imported"`
		NumberCount    int64                 `json:"numberCount" bson:"numberCount"`
		Duplicate      int64                 `json:"duplicate" bson:"duplicate"`
		Invalid        int64                 `json:"invalid" bson:"invalid"`
		LandLine       int64                 `json:"landLine" bson:"landLine"`
		CellPhone      int64                 `json:"cellPhone" bson:"cellPhone"`
		Unknown        int64                 `json:"unknown" bson:"unknown"`
		DNCNumbers     int64                 `json:"dncNumber" bson:"dncNumber"`
		Name           string                `json:"name" bson:"name"`
		FileName       string                `json:"fileName" bson:"fileName"`
		FileS3Path     string                `json:"fileS3Path" bson:"fileS3Path"`

		// options
		RemoveDup      bool `json:"removeDup" bson:"removeDup"`
		Shuffle        bool `json:"shuffle" bson:"shuffle"`
		ScrubDNC       bool `json:"scrubDNC" bson:"scrubDNC"`
		ScrubLandLine  bool `json:"scrubLandLine" bson:"scrubLandLine"`
		ScrubCellPhone bool `json:"scrubCellPhone" bson:"scrubCellPhone"`
		Random         bool `json:"random" bson:"random"`
		AreaCode       bool `json:"areaCode" bson:"areaCode"`

		// ---
		ContactGroups ContactGroupRefs `json:"contactGroups,omitempty" bson:"contactGroups,omitempty"`
		CreateDate    time.Time        `json:"createDate" bson:"createDate"`
		UpdateDate    time.Time        `json:"updateDate" bson:"updateDate"`
	}
)

// data structure to represent list of contact list
type ContactLists []ContactList

// data structure to represent contact group reference
type (
	ContactGroupRef struct {
		ContactGroupId bson.ObjectId `json:"contactGroupId" bson:"contactGroupId"`
	}
)

// data structure to represent contact group reference list
type ContactGroupRefs []ContactGroupRef
