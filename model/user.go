package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

const (
	Super    = 1
	Admin    = 2
	Reseller = 3
	Client   = 4
)

type GetCallCostPayload struct {
	UserId bson.ObjectId
	From   time.Time
	To     time.Time
}

type UserStasPayload struct {
	UserId     bson.ObjectId
	CampaignId bson.ObjectId
	From       time.Time
	To         time.Time
}
type (
	// User represents the structure of our resource
	User struct {
		ID          bson.ObjectId      `json:"id" bson:"_id"`
		FirstName   string             `json:"firstName" bson:"firstName" validate:"required"`
		LastName    string             `json:"lastName" bson:"lastName" validate:"required"`
		CompanyName string             `json:"companyName" bson:"companyName" validate:"required"`
		Address     string             `json:"address" bson:"address"`
		City        string             `json:"city" bson:"city"`
		State       string             `json:"state" bson:"state"`
		Country     string             `json:"country" bson:"country"`
		ZipCode     string             `json:"zipCode" bson:"zipCode"`
		Email       string             `json:"email" bson:"email" validate:"required,email"`
		Phone       string             `json:"phone" bson:"phone" validate:"required"`
		Password    string             `json:"password" bson:"password" validate:"required,min=8"`
		Role        int8               `json:"role,omitempty" bson:"role,omitempty" validate:"required"`
		ParentID    bson.ObjectId      `json:"parentid,omitempty" bson:"parentid,omitempty"`
		Account     Account            `json:"account,omitempty" bson:"account,omitempty"`
		Customize   Customize          `json:"customize,omitempty" bson:"customize,omitempty"`
		Defaults    Defaults           `json:"defaults,omitempty" bson:"defaults,omitempty"`
		IsVerified  bool               `json:"isVerified" bson:"isVerified"`
		IsLocked    bool               `json:"isLocked" bson:"isLocked"`
		UserStats   UserCampaignsStats `json:"userStats" bson:"userStats"`
		TimeSetting TimeSetting        `json:"timesetting" bson:"timesetting"`
		//	Limits          Limits        `json:"limits,omitempty" bson:"limits,omitempty"`
		//Restrictions    *Restrictions `json:"restrictions" bson:"restrictions"`
		APIToken   string    `json:"apiToken" bson:"apiToken"`
		CreateDate time.Time `json:"createDate" bson:"createDate"`
		UpdateDate time.Time `json:"updateDate" bson:"updateDate"`
	}

	Limits struct {
		CallLimit int `json:"callLimit" bson:"callLimit"`
	}

	Account struct {
		BillingType         int8           `json:"billingType,omitempty" bson:"billingType,omitempty"`
		BillingRateText     float32        `json:"billingRateText,omitempty" bson:"billingRateText,omitempty"`
		BillingRateCall     float32        `json:"billingRateCall,omitempty" bson:"billingRateCall,omitempty"`
		BillingRateRingLess float32        `json:"billingRateRingLess,omitempty" bson:"billingRateRingLess,omitempty"`
		HomePageUrl         string         `json:"homePageUrl,omitempty" bson:"homePageUrl,omitempty"`
		Balance             float32        `json:"balance" bson:"balance"`
		IsTrusted           bool           `json:"isTrusted" bson:"isTrusted"`
		Prefix              string         `json:"prefix" bson:"prefix"`
		CarrierId           *bson.ObjectId `json:"carrierId" bson:"carrierId"`
		BlockLandline       bool           `json:"blockLandline" bson:"blockLandline"`
		EnforceDnc          bool           `json:"enforceDnc" bson:"enforceDnc"`
	}
	TimeSetting struct {
		TimeZone string `json:"timezone" bson:"timezone"`
	}

	Customize struct {
		LogoPath       string `json:"logoPath" bson:"logoPath,omitempty"`
		PrimaryColor   string `json:"primaryColor" bson:"primaryColor,omitempty"`
		SecondaryColor string `json:"secondaryColor" bson:"secondaryColor,omitempty"`
	}

	Defaults struct {
		CallRate            float32 `json:"billingRate" bson:"billingRate,omitempty"`
		RinglessRate        float32 `json:"ringlessRate" bson:"ringlessRate,omitempty"`
		InitialPurchaseRate float32 `json:"initialPurchaseRate" bson:"initialPurchaseRate,omitempty"`
		MonthlyPurchaseRate float32 `json:"monthlyPurchaseRate" bson:"monthlyPurchaseRate,omitempty"`
		UseDefaultRates     bool    `json:"useDefaultRates" bson:"useDefaultRates"`
	}

	// structure represents the campaign stats
	UserCampaignsStats struct {
		Total            int64   `json:"total" bson:"total"`
		Dialed           int64   `json:"dialed" bson:"dialed"`
		Busy             int64   `json:"busy" bson:"busy"`
		BusyDuration     int64   `json:"busyDuration" bson:"busyDuration"`
		Error            int64   `json:"error" bson:"error"`
		ErrorDuration    int64   `json:"errorDuration" bson:"errorDuration"`
		NoAnswer         int64   `json:"noAnswer" bson:"noAnswer"`
		NoAnsDuration    int64   `json:"noAnsDuration" bson:"noAnsDuration"`
		Live             int64   `json:"live" bson:"live"`
		LiveDuration     int64   `json:"liveDuration" bson:"liveDuration"`
		VoiceMail        int64   `json:"voiceMail" bson:"voiceMail"`
		VMDuration       int64   `json:"vmDuration" bson:"vmDuration"`
		DNC              int64   `json:"dnc" bson:"dnc"`
		DNCDuration      int64   `json:"dncDuration" bson:"dncDuration"`
		Transfer         int64   `json:"transfer" bson:"transfer"`
		TransferDuration int64   `json:"transferDuration" bson:"transferDuration"`
		Successful       int64   `json:"successful" bson:"successful"`
		LiveFileDuration int64   `json:"liveFileDuration" bson:"liveFileDuration"`
		Avgld            float64 `json:"avgld"`
	}
)

type Users []User
