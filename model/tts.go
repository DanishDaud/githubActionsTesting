package model

import (
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

type TTSDefaultTypeFieldType string

const (
	TTSDefaultFieldTypeFirstName TTSDefaultTypeFieldType = "First Name"
	TTSDefaultFieldTypeLastName  TTSDefaultTypeFieldType = "Last Name"
	TTSDefaultFieldTypeNumber    TTSDefaultTypeFieldType = "Number"
	TTSDefaultFieldTypeAddress   TTSDefaultTypeFieldType = "Address"
	TTSDefaultFieldTypeEmail     TTSDefaultTypeFieldType = "Email"
	TTSDefaultFieldTypeLocation  TTSDefaultTypeFieldType = "Location"
	TTSDefaultFieldTypeCountry   TTSDefaultTypeFieldType = "Country"
	TTSDefaultFieldTypeCity      TTSDefaultTypeFieldType = "City"
	TTSDefaultFieldTypeState     TTSDefaultTypeFieldType = "State"
)

var fieldTypeMap = map[TTSDefaultTypeFieldType]string{
	TTSDefaultFieldTypeFirstName: "first_name",
	TTSDefaultFieldTypeLastName:  "last_name",
	TTSDefaultFieldTypeNumber:    "number",
	TTSDefaultFieldTypeAddress:   "address",
	TTSDefaultFieldTypeEmail:     "email",
	TTSDefaultFieldTypeLocation:  "location",
	TTSDefaultFieldTypeCountry:   "country",
	TTSDefaultFieldTypeCity:      "city",
	TTSDefaultFieldTypeState:     "state",
}

func GetTTSDefaultFieldValueForKey(key TTSDefaultTypeFieldType) string {
	val, ok := fieldTypeMap[key]
	if !ok {
		return ""
	}

	return val
}

type TTSFile struct {
	ID          bson.ObjectId `json:"_id" bson:"_id"`
	FileName    string        `json:"file_name" bson:"file_name"`
	UserID      bson.ObjectId `json:"user_id" bson:"user_id"`
	S3Path      string        `json:"s3_path" bson:"s3_path"`
	CreatedDate time.Time     `json:"created_date" bson:"created_date"`
}

type (
	TTSList struct {
		ID             bson.ObjectId         `json:"id" bson:"_id"`
		UserID         bson.ObjectId         `json:"userid,omitempty" bson:"userid"`
		Status         ContactListStatus     `json:"status" bson:"status"`
		DNCScrubStatus DNCScrubProcessStatus `json:"dncScrubStatus" bson:"dncScrubStatus"`
		FieldData      []Fields              `json:"field_data" bson:"field_data"`
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

		Err string `json:"err" bson:"err"`

		// ---
		ContactGroups ContactGroupRefs `json:"contactGroups,omitempty" bson:"contactGroups,omitempty"`
		CreateDate    time.Time        `json:"createDate" bson:"createDate"`
		UpdateDate    time.Time        `json:"updateDate" bson:"updateDate"`
	}

	DataColumn struct {
		NameCol     int
		LastNameCol int
		NumCol      int
		AddressCol  int
		EmailCol    int
	}
)

type Fields struct {
	Name    string `json:"name" bson:"name"`
	NameKey string `json:"name_key" bson:"name_key"`
}
type ImportData struct {
	FileID         string `json:"file_id"`
	RemoveDup      bool `json:"removeDup" `
	Shuffle        bool `json:"shuffle" `
	ScrubDNC       bool `json:"scrubDNC" `
	ScrubLandLine  bool `json:"scrubLandLine" `
	ScrubCellPhone bool `json:"scrubCellPhone" `
	UserID         string
	Data           struct {
		FirstName    int               `json:"first_name"`
		LastName     int               `json:"last_name"`
		Number       int               `json:"number"`
		Address      int               `json:"address"`
		Email        int               `json:"email"`
		Location     int               `json:"location"`
		Country      int               `json:"country"`
		City         int               `json:"city"`
		State        int               `json:"state"`
		Customfields []CustomFieldMeta `json:"customfields"`
	} `json:"meta"`
}

type CustomFieldMeta struct {
	Name   string `json:"name"`
	Column int    `json:"column"`
}

func (cfm *CustomFieldMeta) NameAndKey() (string, string) {
	name := strings.TrimSpace(cfm.Name)
	key := strings.ToLower(cfm.Name)
	key = strings.TrimSpace(key)
	key = strings.Replace(key, " ", "_", -1)

	return name, key
}
