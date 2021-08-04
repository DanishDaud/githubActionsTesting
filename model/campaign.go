package model

import (
	"strings"
	"time"

	"github.com/gomarkho/sas-rvm-provapi/system/cmlconstants"
	"gopkg.in/mgo.v2/bson"
)

const (
	TransferSettingsItemDNC    = 1
	TransferSettingsItemDirect = 2
	TransferSettingsItemSIP    = 3
)
const (
	All         = 0
	VoiceMail   = 1
	Live        = 2
	Busy        = 3
	NotAnswered = 4
	Failure     = 5
)
const (
	NotDialed    = 1
	Dialed       = 2
	Dialed1time  = 3
	Dialed2times = 4
	Dialed3times = 5
)
const (
	VMDetectBehaviourHangup = 1
	VMDetectBehaviourPlay   = 2
)
const (
	Voicemail      = 5
	LiveStatus     = 1
	BusyStatus     = 2
	Failed         = 4
	NoAnswer       = 3
	NotDialedCalls = 6
)

type (
	// structure represents the campaign stats
	Stats struct {
		Total              int64 `json:"total" bson:"total"`
		Dialed             int64 `json:"dialed" bson:"dialed"`
		Cancel             int64 `json:"cancel" bson:"cancel"`
		CancelDuration     int64 `json:"cancelDuration" bson:"cancelDuration"`
		Busy               int64 `json:"busy" bson:"busy"`
		BusyDuration       int64 `json:"busyDuration" bson:"busyDuration"`
		Error              int64 `json:"error" bson:"error"`
		ErrorDuration      int64 `json:"errorDuration" bson:"errorDuration"`
		Congestion         int64 `json:"congestion" bson:"congestion"`
		CongestionDuration int64 `json:"congestionDuration" bson:"congestionDuration"`
		NoAnswer           int64 `json:"noAnswer" bson:"noAnswer"`
		NoAnsDuration      int64 `json:"noAnsDuration" bson:"noAnsDuration"`
		Live               int64 `json:"live" bson:"live"`
		LiveDuration       int64 `json:"liveDuration" bson:"liveDuration"`
		Answer             int64 `json:"answer" bson:"answer"`
		AnsDuration        int64 `json:"ansDuration" bson:"ansDuration"`
		VoiceMail          int64 `json:"voiceMail" bson:"voiceMail"`
		VMDuration         int64 `json:"vmDuration" bson:"vmDuration"`
		DNC                int64 `json:"dnc" bson:"dnc"`
		DNCDuration        int64 `json:"dncDuration" bson:"dncDuration"`
		Transfer           int64 `json:"transfer" bson:"transfer"`
		TransferAns        int64 `json:"transferAns" bson:"transferAns"`
		TransferDuration   int64 `json:"transferDuration" bson:"transferDuration"`
		TransferTenPlus    int64 `json:"transferTenPlus" bson:"transferTenPlus"`
		LiveAleg           int64 `json:"liveAleg" bson:"liveAleg"`
		//LiveTransfer       int64   `json:"liveTransfer" bson:"liveTransfer"`
		LiveTransfer       int64   `json:"liveTransfer" bson:"liveTransfer"`
		Successful       int64   `json:"successful" bson:"successful"`
		LiveFileDuration int64   `json:"liveFileDuration" bson:"liveFileDuration"`
		Avgld            float64 `json:"avgld"`
		Minutes          int64   `json:"minutes"bson:"minutes"`
	}

	ListeningStats struct {
		Thirty      int64 `json:"thirty" bson:"thirty"`
		Sixty       int64 `json:"sixty" bson:"sixty"`
		FiveMin     int64 `json:"fiveMin" bson:"fiveMin"`
		TenMin      int64 `json:"tenMin" bson:"tenMin"`
		TwentyMin   int64 `json:"twentyMin" bson:"twentyMin"`
		MoreTwentry int64 `json:"moreTwenty" bson:"moreTwenty"`
	}
)

type CampaingList []Campaign

// New Campaign Model
type (
	Campaign struct {
		ID             bson.ObjectId  `json:"id" bson:"_id"`
		Status         int8           `json:"status" bson:"status"`
		Name           string         `json:"name" bson:"name" validate:"required"`
		UserID         bson.ObjectId  `json:"userid,omitempty" bson:"userid"`
		ParentID       bson.ObjectId  `json:"parentid,omitempty" bson:"parentid"`
		CallCampaign   CallCampaign   `json:"call" bson:"call" validate:"required"`
		Error          Error          `json:"error" bson:"error"`
		Cost           Cost           `json:"cost" bson:"cost"`
		Stats          Stats          `json:"stats" bson:"stats"`
		ListeningStats ListeningStats `json:"listeningStats" bson:"listeningStats"`
		CreateDate     time.Time      `json:"createDate" bson:"createDate"`
		UpdateDate     time.Time      `json:"updateDate" bson:"updateDate"`
	}
	CampaignSpeed struct {
		Speed int `json:"speed" bson:"speed" validate:"required"`
	}

	Cost struct {
		CampaignCost float64 `json:"campaignCost" bson:"campaignCost"`
		TransferCost float64 `json:"transferCost" bson:"transferCost"`
	}

	CallCampaign struct {
		ContactListId    bson.ObjectId    `json:"contactlistId" bson:"contactlistId,omitempty"`
		TTSListId        bson.ObjectId    `json:"ttslistId" bson:"ttslistId,omitempty"`
		ContactListName  string           `json:"contactlistName" bson:"contactlistName"`
		DNCListId        bson.ObjectId    `json:"dncListId,omitempty" bson:"dncListId,omitempty"`
		CallerGroupId    *bson.ObjectId   `json:"callerGroupId,omitempty" bson:"callerGroupId,omitempty"`
		CallerId         string           `json:"callerId" bson:"callerId"`
		IsTTS            bool             `json:"isTTS" bson:"isTTS"`
		TTSText          string           `json:"ttstext" bson:"ttstext"`
		SoundFileId      *bson.ObjectId   `json:"soundFileId,omitempty" bson:"soundFileId,omitempty"`
		Speed            int              `json:"speed" bson:"speed" validate:"required"`
		MCD              int              `json:"mcd" bson:"mcd" validate:"required"`
		RetrySettings    RetrySettings    `json:"retrySettings" bson:"retrySettings"`
		ScheduleSettings ScheduleSettings `json:"scheduleSettings" bson:"scheduleSettings"`
	}

	TransferSettings struct {
		Items []TransferSettingsItem `json:"items" bson:"items"`
	}

	TransferSettingsItem struct {
		Key   string         `json:"key" bson:"key"`
		Type  int            `json:"type" bson:"type"`
		DncId *bson.ObjectId `json:"dncId,omitempty" bson:"dncId,omitempty"`
		Info  string         `json:"info" bson:"info"`
	}

	RetrySettings struct {
		Retry        int  `json:"retry" bson:"retry"`
		Delay        int  `json:"delay" bson:"delay"`
		ToCompletion bool `json:"toCompletion" bson:"toCompletion"`
	}

	BehaviorItem struct {
		Type int `json:"type" bson:"type"`
	}

	VoiceMailSettings struct {
		VMDetect      bool           `json:"vmDetect" bson:"vmDetect"`
		Type          int            `json:"type" bson:"type"`
		VMSoundFileId *bson.ObjectId `json:"vmSoundFileId,omitempty" bson:"vmSoundFileId,omitempty"`
	}

	ScheduleSettings struct {
		Id string `json:"id" bson:"id"`
		// start info
		//StartDate   *time.Time `json:"startDate" bson:"startDate"`
		StartDateUI time.Time `json:"startDateUI, omitempty" validate:"required"`
		//	StartTime   []int      `json:"startTime" bson:"startTime" validate:"required"`
		// end info
		//EndDate   *time.Time `json:"endDate" bson:"endDate"`
		EndDateUI time.Time `json:"endDateUI, omitempty" validate:"required"`
		//	EndTime   []int      `json:"endTime" bson:"endTime" validate:"required"`
		// others
		TimeZones         []string `json:"timeZones" bson:"timeZones" validate:"required"`
		IncludeDays       []int    `json:"includeDays" bson:"includeDays" validate:"required"`
		RangeStartTime    []int    `json:"rangeStartTime" bson:"rangeStartTime" validate:"required"`
		RangeEndTime      []int    `json:"rangeEndTime" bson:"rangeEndTime" validate:"required"`
		ScheduleStartTime []int    `json:"scheduleStartTime" bson:"scheduleStartTime"`
		ScheduleEndTime   []int    `json:"scheduleEndTime" bson:"scheduleEndTime"`
		Stopped           bool     `json:"stopped" bson:"stopped"`
	}

	Restrictions struct {
		StartTime []int  `json:"startTime" bson:"startTime" validate:"required"`
		EndTime   []int  `json:"endTime" bson:"endTime" validate:"required"`
		TimeZone  string `json:"timeZone" bson:"timeZone" validate:"required"`
	}
)

func (s *ScheduleSettings) IncludeDaysSystem() []time.Weekday {
	var days []time.Weekday
	for _, i := range s.IncludeDays {
		switch i {
		case 1:
			days = append(days, time.Monday)
		case 2:
			days = append(days, time.Tuesday)
		case 3:
			days = append(days, time.Wednesday)
		case 4:
			days = append(days, time.Thursday)
		case 5:
			days = append(days, time.Friday)
		case 6:
			days = append(days, time.Saturday)
		case 7:
			days = append(days, time.Sunday)
		}
	}

	return days
}

type DuplicateCampaign struct {
	Name          string `json:"name" validate:"required"`
	ContactListId string `json:"contactlistId" bson:"contactlistId"`
	TTSListID     string `json:"ttslistid" bson:"ttslistid"`
}

type CallLiveStats struct {
	Live     int
	Transfer int
}

func (cm *Campaign) CanBeDeleted() bool {
	if cm.Status == cmlconstants.CampaignStatusNew || cm.Status == cmlconstants.CampaignStatusStopped || cm.Status == cmlconstants.CampaignStatusFinished {
		return true
	}

	return false
}

func (cm *Campaign) TimeZoneIncluded(tz TimeZone) bool {
	for _, t := range cm.CallCampaign.ScheduleSettings.TimeZones {
		if strings.ToLower(t) == strings.ToLower(string(tz)) {
			return true
		}
	}

	return false
}
