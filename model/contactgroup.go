package model

import (
	"strings"
	"time"

	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

type TimeZone string
type TimeZoneCode string

const (
	CST TimeZone = "US/Central"
	PST          = "US/Pacific"
	EST          = "US/Eastern"
	MST          = "US/Mountain"
)

const (
	CSTCode TimeZoneCode = "CST"
	PSTCode              = "PST"
	ESTCode              = "EST"
	MSTCode              = "MST"
)

var TimeZoneMap = map[TimeZoneCode]TimeZone{
	CSTCode: CST,
	PSTCode: PST,
	ESTCode: EST,
	MSTCode: MST,
}

type (
	// User represents the structure of our resource
	ContactGroup struct {
		ID            bson.ObjectId `json:"id" bson:"_id"`
		ContactListId bson.ObjectId `json:"contactListId" bson:"contactListId"`
		Numbers       NumberList    `json:"numbers,omitempty" bson:"numbers,omitempty"`
		CreateDate    time.Time     `json:"createDate" bson:"createDate"`
		UpdateDate    time.Time     `json:"updateDate" bson:"updateDate"`
		TotalNumbers  int32         `json:"totalNumbers" bson:"totalNumbers"`
	}

	TTSListContactGroup struct {
		ID           bson.ObjectId `json:"id" bson:"_id"`
		TTSListId    bson.ObjectId `json:"contactListId" bson:"contactListId"`
		Numbers      TTSNumberList `json:"numbers,omitempty" bson:"numbers,omitempty"`
		CreateDate   time.Time     `json:"createDate" bson:"createDate"`
		UpdateDate   time.Time     `json:"updateDate" bson:"updateDate"`
		TotalNumbers int32         `json:"totalNumbers" bson:"totalNumbers"`
	}
)

type ContactGroups []ContactGroup

// dataStructure for number

type NumberType int8

const (
	NumberTypeCellPhone NumberType = 1
	NumberTypeLandLine             = 2
	NumberTypeUnknown              = 3
)

type (
	Number struct {
		ID            bson.ObjectId `json:"id" bson:"_id"`
		Number        string        `json:"number" bson:"number"`
		TimeZone      string        `json:"timezone" bson:"timezone"`
		NumberTypeStr string        `json:"numbertype" bson:"numbertype"`
		Text          string        `json:"text" bson:"text"`
		NumberType    NumberType    `json:"numbertypeval" bson:"numbertypeval"`
	}
)
type NumberList []Number

type (
	TTSMetaNumber struct {
		ID            bson.ObjectId  `json:"id" bson:"_id"`
		TTSListID     bson.ObjectId  `json:"ttslistID" bson:"ttslistID"`
		Number        string         `json:"number" bson:"number"`
		TimeZone      string         `json:"timezone" bson:"timezone"`
		NumberTypeStr string         `json:"numbertype" bson:"numbertype"`
		NumberType    NumberType     `json:"numbertypeval" bson:"numbertypeval"`
		FieldData     []TTSFieldData `json:"customdata" bson:"customdata"`
	}

	TTSNumber struct {
		ID            bson.ObjectId  `json:"id" bson:"_id"`
		Number        string         `json:"number" bson:"number"`
		TimeZone      string         `json:"timezone" bson:"timezone"`
		NumberTypeStr string         `json:"numbertype" bson:"numbertype"`
		NumberType    NumberType     `json:"numbertypeval" bson:"numbertypeval"`
		FieldData     []TTSFieldData `json:"customdata" bson:"customdata"`
	}

	TTSFieldData struct {
		Name    string `json:"name" bson:"name"`
		NameKey string `json:"name_key" bson:"name_key"`
		Data    string `json:"data" bson:"data"`
	}
)

type TTSNumberList []TTSNumber

type TTSMetaList []TTSMetaNumber

func (num *Number) HasValidTimeZone() bool {
	_, ok := TimeZoneMap[TimeZoneCode(num.TimeZone)]
	return ok
}

func (num *Number) IsValidCallerId() bool {
	var flag bool

	number := strings.ToLower(num.Number)

	if number != "" {
		if strings.Contains(number, ",") {
			sArray := strings.Split(number, ",")
			for _, i := range sArray {
				callerid := strings.TrimSpace(i)
				if callerid == "" {
					continue
				}
				if strings.Contains(callerid, "x") {
					f2 := i[:2]
					if strings.Contains(f2, "x") == true {
						flag = false
						return flag

					}
					count := strings.Count(callerid, "x")
					n := strings.Replace(callerid, "x", "4", count)
					flag = cmlutils.CallerIDVerification(n)
				} else {
					flag = cmlutils.CallerIDVerification(callerid)
				}
				if flag == true {
					continue
				} else {
					return flag
				}
			}
		} else if strings.Contains(number, "x") {
			number = strings.TrimSpace(number)
			f2 := number[:2]
			if strings.Contains(f2, "x") == true {
				flag = false
				return flag

			}
			count := strings.Count(number, "x")
			n := strings.Replace(number, "x", "4", count)
			flag = cmlutils.CallerIDVerification(n)
			return flag

		} else {
			flag = cmlutils.CallerIDVerification(number)
			return flag
		}
	}
	return true
}

func (num *TTSNumber) HasValidTimeZone() bool {
	_, ok := TimeZoneMap[TimeZoneCode(num.TimeZone)]
	return ok
}

func (num *TTSNumber) IsValidCallerId() bool {
	var flag bool

	number := strings.ToLower(num.Number)

	if number != "" {
		if strings.Contains(number, ",") {
			sArray := strings.Split(number, ",")
			for _, i := range sArray {
				callerid := strings.TrimSpace(i)
				if callerid == "" {
					continue
				}
				if strings.Contains(callerid, "x") {
					f2 := i[:2]
					if strings.Contains(f2, "x") == true {
						flag = false
						return flag

					}
					count := strings.Count(callerid, "x")
					n := strings.Replace(callerid, "x", "4", count)
					flag = cmlutils.CallerIDVerification(n)
				} else {
					flag = cmlutils.CallerIDVerification(callerid)
				}
				if flag == true {
					continue
				} else {
					return flag
				}
			}
		} else if strings.Contains(number, "x") {
			number = strings.TrimSpace(number)
			f2 := number[:2]
			if strings.Contains(f2, "x") == true {
				flag = false
				return flag

			}
			count := strings.Count(number, "x")
			n := strings.Replace(number, "x", "4", count)
			flag = cmlutils.CallerIDVerification(n)
			return flag

		} else {
			flag = cmlutils.CallerIDVerification(number)
			return flag
		}
	}
	return true
}

//US/Central - CST
//US/Pacific - PST
//US/Eastern - EST
//US/Mountain - MST
