package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type UserStats struct {
	ActiveCampaigns int     `json:"activecampaigns"`
	ActiveCall      int     `json:"activecall"`
	ActiveTransfer  int     `json:"activetransfer"`
	Cost            float32 `json:"cost"`
}

type CallReport struct {
	AlegInformation       Aleg           `json:"aleginforamtion"`
	BlegInforamtion       Bleg           `json:"bleginforamtion"`
	CallLengthInformation ListeningStats `json:"calllengthinformation"`
}
type Aleg struct {
	UserID       bson.ObjectId `json:"userid,omitempty" bson:"userid"`
	CampaignID   bson.ObjectId `json:"campaignid,omitempty" bson:"campaignid"`
	Dail         int           `json:"dail" bson:"dail"`
	Redail       int           `json:"redail"`
	BillDuration int           `json:"billduration" bson:"billdur"`
	IncrDur      int           `json:"incrdur" bson:"incrdur"`
	Cost         float32       `json:"cost" bson:"cost"`
	Answer       int           `json:"answer" bson:"answer"`
	NoAnswer     int           `json:"noanswer" bson:"no-answer"`
	Live         int           `json:"live" bson:"live"`
	VM           int           `json:"vm" bson:"vm"`
	Busy         int           `json:"busy" bson:"busy"`
	Failed       int           `json:"failed" bson:"error"`
	Aloc         float32       `json:"alo"`
	ASR          float32       `json:"asr"`
	Liveper      float32       `json:"liveper"`
	Vmper        float32       `json:"vmper"`
	Xfer         float32       `json:"xfer"`
	DNC          int           `json:"dnc" bson:"dnc"`
	Timeout      int           `json:"time_out" bson:"time_out"`
	CreatedDate  time.Time     `json:"-" bson:"created_date"`
}

type Bleg struct {
	UserID       bson.ObjectId `json:"userid,omitempty" bson:"userid"`
	CampaignID   bson.ObjectId `json:"campaignid,omitempty" bson:"campaignid"`
	Dail         int           `json:"dail" bson:"dail"`
	BillDuration int           `json:"billduration" bson:"billdur"`
	IncrDur      int           `json:"incrdur" bson:"incrdur"`
	Cost         float32       `json:"cost" bson:"cost"`
	Aloc         float32       `json:"alo"`
	ASR          float32       `json:"asr"`
	Answer       int           `json:"answer" bson:"answer"`
	NoAnswer     int           `json:"noanswer" bson:"no-answer"`
	Thirty       int           `json:"-" bson:"thirty"`
	Sixty        int           `json:"-" bson:"sixty"`
	FiveMin      int           `json:"-" bson:"five_min"`
	TenMin       int           `json:"-" bson:"ten_min"`
	TwentyMin    int           `json:"-" bson:"twenty_min"`
	MoreTwenty   int           `json:"-" bson:"more_twenty"`
	CreatedDate  time.Time     `json:"-" bson:"created_date"`
}
type AlegHourly struct {
	To           int           `json:"to" bson:"to"`
	Form         int           `json:"form" bson:"from"`
	UserID       bson.ObjectId `json:"userid,omitempty" bson:"userid"`
	CampaignID   bson.ObjectId `json:"campaignid,omitempty" bson:"campaignid"`
	Dail         int           `json:"dail" bson:"dail"`
	Redail       int           `json:"redail"`
	BillDuration int           `json:"billduration" bson:"billdur"`
	IncrDur      int           `json:"incrdur" bson:"incrdur"`
	Cost         float32       `json:"cost" bson:"cost"`
	Answer       int           `json:"answer" bson:"answer"`
	NoAnswer     int           `json:"noanswer" bson:"no-answer"`
	Live         int           `json:"live" bson:"live"`
	VM           int           `json:"vm" bson:"vm"`
	Xfer         int           `json:"xfer" bson:"xfer"`
	Busy         int           `json:"busy" bson:"busy"`
	Failed       int           `json:"failed" bson:"error"`
	DNC          int           `json:"dnc" bson:"dnc"`
	Timeout      int           `json:"time_out" bson:"time_out"`
	CreatedDate  time.Time     `json:"created_date" bson:"created_date"`
}
type CallDispositions struct {
	Time     string `json:"time"`
	Answered int    `json:"answered" bson:"answer"`
	Busy     int    `json:"busy" bson:"busy"`
	NoAnswer int    `json:"noanswer" bson:"no_answer"`
	Failed   int    `json:"failed" bson:"error"`
	DNC      int    `json:"dnc" bson:"dnc"`
	VM       int    `json:"voicemail" bson:"vm"`
}

type CallInformation struct {
	Userid       string    `json:"userid" bson:"userid"`
	To           int       `json:"to" bson:"to"`
	Form         int       `json:"form" bson:"from"`
	Answerd      int       `json:"answerd" bson:"answer"`
	Busy         int       `json:"busy" bson:"busy"`
	NoAnswer     int       `json:"no_answer" bson:"no-answer"`
	Failed       int       `json:"failed" bson:"error"`
	DNC          int       `json:"dnc" bson:"dnc"`
	Timeout      int       `json:"timeout" bson:"time_out"`
	BillDuration int       `json:"bill_duration" bson:"billdur"`
	Cost         float32   `json:"cost" bson:"cost"`
	AlegCount    int       `json:"aleg" bson:"aleg"`
	BlegCount    int       `json:"bleg " bson:"bleg"`
	VM           int       `json:"voiceMail" bson:"vm"`
	CreatedDate  time.Time `json:"created_date" bson:"created_date" `
}
type CallLoc struct {
	Time string  `json:"time"`
	ASR  float64 `json:"asr"`
	ALOC float64 `json:"aloc"`
}

type CallCount struct {
	Time  string `json:"time"`
	Min   int    `json:"min"`
	Count int    `json:"count"`
}
type CallCost struct {
	Time string  `json:"time"`
	Cost float32 `json:"cost"`
}

type ALegCall struct {
	Time  string `json:"time"`
	Count int    `json:"count"`
}

type FromNumber struct {
	Number string `json:"number" bson:"callTo"`
}
