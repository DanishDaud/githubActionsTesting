package service

import (
	"encoding/json"
	"fmt"

	"io/ioutil"
	"net/http"

	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
)

type EventsService struct {
	Service
}

type EventsStatsResponse struct {
	Data EventStats `json:"data"`
}

type EventStats struct {
	RingingCount  int `json:"ringing_count"`
	TransferCount int `json:"transfer_count"`
	AnswerCount   int `json:"answer_count"`
}

func (es *EventsService) GetUserStats(id string) *EventStats {
	var stats EventStats

	url := fmt.Sprintf(cmlutils.EventsAPIPath()+"user/%s/live/stats", id)
	resp, err := http.Get(url)
	if err != nil {
		return &stats
	}

	defer resp.Body.Close()
	body, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		return &stats
	}

	// Unmarshal
	var msg EventsStatsResponse
	err = json.Unmarshal(body, &msg)
	if err != nil {
		return &stats
	}

	stats = msg.Data

	return &stats
}

func (es *EventsService) GetCampaignStats(id string) *EventStats {
	var stats EventStats

	url := fmt.Sprintf(cmlutils.EventsAPIPath()+"campaign/%s/live/stats", id)
	resp, err := http.Get(url)
	if err != nil {
		return &stats
	}

	defer resp.Body.Close()
	body, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		return &stats
	}

	// Unmarshal
	var msg EventsStatsResponse
	err = json.Unmarshal(body, &msg)
	if err != nil {
		return &stats
	}

	stats = msg.Data

	return &stats
}
