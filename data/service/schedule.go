package service

/*import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"

)

type ScheduleService struct {
}

func NewScheduleService() *ScheduleService {
	return &ScheduleService{}
}

type ScheduleRequest struct {
	Id string `json:"id" bson:"id" validate:"required"`

	// schedule
	ResumeNextDay bool  `json:"resumeNextDay" bson:"resumeNextDay"`
	StartTime     []int `json:"startTime" bson:"startTime"`
	IncludeDays   []int `json:"includeDays" bson:"includeDays"`

	// time restrictions
	RangeStartTime []int  `json:"rangeStartTime" bson:"rangeStartTime"`
	RangeEndTime   []int  `json:"rangeEndTime" bson:"rangeEndTime"`
	TimeZone       string `json:"timeZone" bson:"timeZone" validate:"required"`

	// only for processing
	StartDateUI string `json:"startDateUI, omitempty"`
}

func (ss *ScheduleService) Update(cam model.Campaign) error {
	url := fmt.Sprintf(cmlutils.VOIPAPIPath() + "user/%s/campaign/%s/schedule", cam.UserID.Hex(), cam.ID.Hex())
	clnt := &http.Client{}

	sch := cam.CallCampaign.Schedule
	tr := cam.CallCampaign.Settings.Restrictions

	body := ScheduleRequest{
		Id:             sch.Id,
		ResumeNextDay:  sch.ResumeNextDay,
		StartTime:      sch.StartTime,
		IncludeDays:    sch.IncludeDays,
		RangeStartTime: tr.StartTime,
		RangeEndTime:   tr.EndTime,
		TimeZone:       tr.TimeZone,
		StartDateUI:    sch.StartDateUI,
	}

	bts, err := json.Marshal(&body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(bts))
	req.Header.Add("Content-type", "application/json")
	resp, err := clnt.Do(req)
	if err != nil {
		return err
	}

	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
		return errors.New("Schedule save failed")
	}

	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (ss *ScheduleService) Delete(cam *model.Campaign) error {
	url := fmt.Sprintf(cmlutils.VOIPAPIPath() +"user/%s/campaign/%s/schedule", cam.UserID.Hex(), cam.ID.Hex())

	clnt := &http.Client{}

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	req.Header.Add("Content-type", "application/json")
	resp, err := clnt.Do(req)
	if err != nil {
		return err
	}

	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
		return errors.New("Schedule save failed")
	}

	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (ss *ScheduleService) Create(cam model.Campaign) error {
	url := fmt.Sprintf(cmlutils.VOIPAPIPath() +"user/%s/campaign/%s/schedule", cam.UserID.Hex(), cam.ID.Hex())

	clnt := &http.Client{}

	sch := cam.CallCampaign.Schedule
	tr := cam.CallCampaign.Settings.Restrictions

	body := ScheduleRequest{
		Id:             sch.Id,
		ResumeNextDay:  sch.ResumeNextDay,
		StartTime:      sch.StartTime,
		IncludeDays:    sch.IncludeDays,
		RangeStartTime: tr.StartTime,
		RangeEndTime:   tr.EndTime,
		TimeZone:       tr.TimeZone,
		StartDateUI:    sch.StartDateUI,
	}

	bts, err := json.Marshal(&body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bts))
	req.Header.Add("Content-type", "application/json")
	resp, err := clnt.Do(req)
	if err != nil {
		return err
	}

	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
		return errors.New("Schedule save failed")
	}

	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}*/
