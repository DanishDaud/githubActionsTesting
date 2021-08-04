package model

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	AMDStatusPerson = 1
	AMDStatusVM     = 2
)

const (
	DispositionAnswer        = 1
	DispositionBusy          = 2
	DispositionNoAnswer      = 3
	DispositionError         = 4
	DispositionDNC           = 5
	DispositionAnswerTimeOut = 6
)

const (
	CallTypeALeg = 1
	CallTypeBLeg = 2
)
const (
	Last24     = 3
	OneHour    = 5
	ThreeHour  = 4
	Thirtydays = 1
	SevenDays  = 2
)

type CDR struct {
	ID              bson.ObjectId `json:"id" bson:"_id"`
	UserID          bson.ObjectId `json:"userid,omitempty" bson:"userid"`
	ParentID        bson.ObjectId `json:"parentid,omitempty" bson:"parentid"`
	BillingParentID bson.ObjectId `json:"billingParentid,omitempty" bson:"billingParentid"`
	AMDStatus       int8          `json:"amdStatus" bson:"amdStatus"`
	Disposition     int8          `json:"disposition" bson:"disposition"`
	CallType        int           `json:"calltype" bson:"calltype"`
	CampaignId      bson.ObjectId `json:"campaignId" bson:"campaignId"`
	IsTestCall      bool          `json:"isTestCall" bson:"isTestCall"`
	SipStatus       string        `json:"sip_status" bson:"sip_status"`
	SipCode         string        `json:"sip_code" bson:"sip_code"`
	CampaignName    string        `json:"campaignName" bson:"campaignName"`
	CallTo          string        `json:"callTo" bson:"callTo"`
	CallFrom        string        `json:"callFrom" bson:"callFrom"`
	BillingIncr     int           `json:"billIncr" bson:"billIncr"`
	CallerId        string        `json:"callerId" bson:"callerId"`
	CallTime        int           `json:"callTime" bson:"callTime"`
	BillDuration    int           `json:"billDuration" bson:"billDuration"`
	StartDate       time.Time     `json:"startDate" bson:"startDate"`
	Cost            float32       `json:"cost" bson:"cost"`
	CreateDate      time.Time     `json:"createDate" bson:"createDate"`
	UpdateDate      time.Time     `json:"updateDate" bson:"updateDate"`
}

type DailyCDR struct {
	Date              string  `json:"date"`
	Call              int     `json:"calls"`
	Answer            int     `json:"answer"`
	BillDuration      int     `json:"billduration"`
	IncrementDuration int     `json:"incrementduration"`
	ALOC              int     `json:"aloc"`
	Cost              float64 `json:"cost"`
}
type HourlyCDR struct {
	To                int     `json:"to"`
	From              int     `json:"from"`
	TotalCall         int     `json:"totalcall"`
	Answer            int     `json:"answer"`
	Busy              int     `json:"busy"`
	NoAnswer          int     `json:"noanswer"`
	Failed            int     `json:"failed"`
	VoiceMail         int     `json:"voicemail"`
	BillDuration      int     `json:"billduration"`
	IncrementDuration int     `json:"incrementduration"`
	ASR               float64 `json:"asr"`
	ALOC              int     `json:"aloc"`
	Cost              float64 `json:"cost"`
}
type CompareCampaign struct {
	Ids []string `json:"id" validate:"required"`
}

type CampaignCDR struct {
	CampaignID     string         `json:"campaignid"`
	CampaignName   string         `json:"campaignname"`
	Contact        int            `json:"contact"`
	Minutes        int            `json:"minutes"`
	Cost           Cost           `json:"cost" bson:"cost"`
	ASR            float64        `json:"asr"`
	ALOC           int            `json:"aloc"`
	Dial           int            `json:"dial"`
	Redial         int            `json:"redial"`
	Stats          Stats          `json:"stats"`
	ListeningStats ListeningStats `json:"listeningStats"`
	Answer         int            `json:"answer"`
	NoAnswer       int            `json:"noanswer"`
	Live           int            `json:"live"`
	LivePercent    float64        `json:"livePerc"`
	Vm             int            `json:"vm"`
	VmPercent      float64        `json:"vmPerc"`
	XAns           int            `json:"xans"`
	XAnsPerc       float64        `json:"xansPerc"`
	CPX            float64        `json:"cpx"`
	Progress       float64        `json:"progress"`
}

type CampaignHourly struct {
	CampaignID string  `json:"campaignid"`
	To         int     `json:"to"`
	From       int     `json:"from"`
	Call       int     `json:"call"`
	Answer     int     `json:"answer"`
	VM         int     `json:"vm"`
	Busy       int     `json:"busy"`
	NoAns      int     `json:"noans"`
	Failed     int     `json:"failed"`
	DNC        int     `json:"dnc"`
	UnAutd     int     `json:"unautd"`
	Minute     int     `json:"minute"`
	ASR        float64 `json:"asr"`
	ALOC       int     `json:"aloc"`
	Cost       float64 `json:"cost"`
	XFer       int     `json:"xfer"`
	XAns       int     `json:"xans"`
	XMin       int     `json:"xmin"`
	XASR       float64 `json:"xasr"`
	XALOC      int     `json:"xaloc"`
	XFerTC     float64 `json:"xfertc"`
	CPX        float64 `json:"cpx"`
}

type CampaignDaily struct {
	Date         string  `json:"date"`
	CampaignID   string  `json:"campaignid"`
	CampaignName string  `json:"campaignname"`
	Created      string  `json:"created"`
	Progress     float64 `json:"progress"`
	Calls        int     `json:"calls"`
	Minute       int     `json:"minute"`
	Cost         float64 `json:"cost"`
	Xfer         int     `json:"xfer"`
	XferMin      int     `json:"xfermin"`
	XferCost     float64 `json:"xfercost"`
	CPX          float64 `json:"cpx"`
}

type DailyList []DailyCDR
type CdrList []CDR
type HourlyCdrList []HourlyCDR
type CampaignCdrList []CampaignCDR
type CampaignDailyList []CampaignDaily
type CampaignHourlyList []CampaignHourly
