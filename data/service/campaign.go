package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
)

// Data Structure to represent campaign service
type CampaignService struct {
	Service
	RedisService
}

type CampaignStat struct {
	ID       bson.ObjectId `json:"id"`
	Live     int           `json:"live"`
	Transfer int           `json:"transfer"`
}
type CallLiveStatsResponse struct {
	Data model.CallLiveStats `json:"data"`
}

func (cms *CampaignService) CampaignLiveStats(campaignId string) (stats model.CallLiveStats) {
	url := fmt.Sprintf(cmlutils.VOIPAPIPath()+"campaign/%s/livestats", campaignId)
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		return
	}

	// Unmarshal
	var msg CallLiveStatsResponse
	err = json.Unmarshal(body, &msg)
	if err != nil {
		return
	}

	stats.Live = msg.Data.Live
	stats.Transfer = msg.Data.Transfer

	return stats
}

func (cms *CampaignService) RunningCampaignsCountForUser(userId string) (int, error) {
	if !bson.IsObjectIdHex(userId) {
		return 0, errors.New("user id is not valid")
	}
	userObjectId := bson.ObjectIdHex(userId)

	// get new instance of campaign data source
	campaignDataSource := cms.campaignDatasource()
	defer campaignDataSource.Session.Close()

	return campaignDataSource.RunningCampaignsCountForUser(userObjectId)
}

func (cms *CampaignService) IsContactListAttached(userId string, objectID string) (bool, error) {
	if !bson.IsObjectIdHex(objectID) || !bson.IsObjectIdHex(userId) {
		return false, errors.New("object id is not valid")
	}

	contactListObjectId := bson.ObjectIdHex(objectID)
	userObjectId := bson.ObjectIdHex(userId)

	// get new instance of campaign data source
	campaignDataSource := cms.campaignDatasource()
	defer campaignDataSource.Session.Close()

	count, err := campaignDataSource.CountWithContactList(userObjectId, contactListObjectId)
	if err != nil {
		return false, errors.New("error while fetching attached sound file count")
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

func (cms *CampaignService) IsTTSListAttached(userId string, objectID string) (bool, error) {
	if !bson.IsObjectIdHex(objectID) || !bson.IsObjectIdHex(userId) {
		return false, errors.New("object id is not valid")
	}

	contactListObjectId := bson.ObjectIdHex(objectID)
	userObjectId := bson.ObjectIdHex(userId)

	// get new instance of campaign data source
	campaignDataSource := cms.campaignDatasource()
	defer campaignDataSource.Session.Close()

	count, err := campaignDataSource.CountWithTTSList(userObjectId, contactListObjectId)
	if err != nil {
		return false, errors.New("error while fetching attached sound file count")
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

func (cms *CampaignService) IsSoundFileAttached(userId string, objectID string) (bool, error) {
	if !bson.IsObjectIdHex(objectID) {
		return false, errors.New("object id is not valid")
	}

	soundFileObjectId := bson.ObjectIdHex(objectID)
	userObjectId := bson.ObjectIdHex(userId)

	// get new instance of campaign data source
	campaignDataSource := cms.campaignDatasource()
	defer campaignDataSource.Session.Close()

	count1, err1 := campaignDataSource.CountWithLiveSoundFileId(userObjectId, soundFileObjectId)
	count2, err2 := campaignDataSource.CountWithVMSoundFileId(userObjectId, soundFileObjectId)
	//count3, err3 := campaignDataSource.CountWithDNCSoundFileId(userObjectId, soundFileObjectId)
	//count4, err4 := campaignDataSource.CountWithTransferSoundFileId(userObjectId, soundFileObjectId)

	if err1 != nil || err2 != nil {
		return false, errors.New("error while fetching attached sound file count")
	}

	if count1 > 0 || count2 > 0 {
		return true, nil
	}

	return false, nil
}

func (cms *CampaignService) CampaignObjectWithId(objectID string) (*model.Campaign, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, errors.New("object id is not valid")
	}

	campaignObjectId := bson.ObjectIdHex(objectID)

	// get new instance of campaign datasource
	campaignDataSource := cms.campaignDatasource()
	defer campaignDataSource.Session.Close()

	return campaignDataSource.CampaignWithId(campaignObjectId)
}

func (cms *CampaignService) DeleteWithId(objectID string) error {
	if !bson.IsObjectIdHex(objectID) {
		return errors.New("object id is not valid")
	}

	campaignObjectId := bson.ObjectIdHex(objectID)

	// get new instance of campaign datasource
	campaignDataSource := cms.campaignDatasource()
	defer campaignDataSource.Session.Close()

	return campaignDataSource.DeleteWithId(campaignObjectId)
}

func (cms *CampaignService) UpdateName(objectID string, name string) error {
	if !bson.IsObjectIdHex(objectID) {
		return errors.New("object id is not valid")
	}

	campaignObjectId := bson.ObjectIdHex(objectID)

	// get new instance of campaign datasource
	campaignDataSource := cms.campaignDatasource()
	defer campaignDataSource.Session.Close()

	return campaignDataSource.UpdateName(campaignObjectId, name)
}

func (cms *CampaignService) UpdateContactListInfo(objectID string, clId string, total int64) error {
	if !bson.IsObjectIdHex(objectID) {
		return errors.New("object id is not valid")
	}

	if !bson.IsObjectIdHex(clId) {
		return errors.New("contact object id is not valid")
	}

	campaignObjectId := bson.ObjectIdHex(objectID)
	contactListObjectId := bson.ObjectIdHex(clId)

	// get new instance of campaign datasource
	campaignDataSource := cms.campaignDatasource()
	defer campaignDataSource.Session.Close()

	return campaignDataSource.UpdateContactListinfo(campaignObjectId, contactListObjectId, total)
}

func (cms *CampaignService) UpdateLimits(objectID string, cpm int, transfer int) error {
	if !bson.IsObjectIdHex(objectID) {
		return errors.New("object id is not valid")
	}

	campaignObjectId := bson.ObjectIdHex(objectID)

	// get new instance of campaign datasource
	campaignDataSource := cms.campaignDatasource()
	defer campaignDataSource.Session.Close()

	return campaignDataSource.UpdateLimits(campaignObjectId, cpm, transfer)
}

func (cms *CampaignService) CampaignObjectWithCampaignIdAndUserId(campaignId string, userId string) (*model.Campaign, error) {
	// check validation of campaign id
	if !bson.IsObjectIdHex(campaignId) {
		return nil, errors.New("campaign id is not valid")
	}
	campaignObjectId := bson.ObjectIdHex(campaignId)

	if !bson.IsObjectIdHex(userId) {
		return nil, errors.New("user id is not valid")
	}
	userObjectId := bson.ObjectIdHex(userId)

	// get new instance of campaign datasource
	campaignDataSource := cms.campaignDatasource()
	defer campaignDataSource.Session.Close()

	return campaignDataSource.CampaignWithCampaignIdAndUserId(campaignObjectId, userObjectId)
}

// this method saves campaign to the database
// through campaign datasource
func (cms *CampaignService) SaveCampaign(campaign *model.Campaign) error {
	// get new instance of sound file datasource
	campaignDataSource := cms.campaignDatasource()
	defer campaignDataSource.Session.Close()

	return campaignDataSource.SaveCampaign(campaign)
}

func (cms *CampaignService) CampaignObjectWithContactId(objectID string) (model.CampaingList, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, errors.New("object id is not valid")
	}

	campaignObjectId := bson.ObjectIdHex(objectID)

	// get new instance of campaign datasource
	campaignDataSource := cms.campaignDatasource()
	defer campaignDataSource.Session.Close()

	return campaignDataSource.CampaignWithContactId(campaignObjectId)
}
func (cms *CampaignService) UpdateCampaignStatus(campaign *model.Campaign, status int8) error {
	// get new instance of sound file datasource
	campaignDataSource := cms.campaignDatasource()
	defer campaignDataSource.Session.Close()

	return campaignDataSource.UpdateCampaignStatus(campaign, status)
}

func (cms *CampaignService) UpdateCampaignSchedule(campaign *model.Campaign, id string) error {
	// get new instance of sound file datasource
	campaignDataSource := cms.campaignDatasource()
	defer campaignDataSource.Session.Close()

	return campaignDataSource.UpdateSchedule(campaign, id)
}

func (cms *CampaignService) CampaignListListingWithUserId(objectID string, page int, limit int) (model.CampaingList, int, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, 0, errors.New("User object id is not valid")
	}

	userObjectId := bson.ObjectIdHex(objectID)

	// get new instance of campaign list datasource
	campaignDataSource := cms.campaignDatasource()
	defer campaignDataSource.Session.Close()

	list, total, err := campaignDataSource.CampaignListGetList(userObjectId, page, limit)
	if err != nil {
		return nil, total, err
	}
	return *list , total , nil

}
func (cms *CampaignService) CampaignListStat(list []string) (*[]CampaignStat, error) {

	logrus.Infoln(list)
	var data CampaignStat
	var result []CampaignStat

	for _, c := range list {

		stats := cms.CampaignLiveStats(c)
		data.ID = bson.ObjectIdHex(c)
		data.Live = (stats.Live)
		data.Transfer = (stats.Transfer)
		result = append(result, data)
	}

	return &result, nil
}

// return instance of campaign  datasource
// every time a new instance would be created
func (cms *CampaignService) campaignDatasource() *datasource.CampaignDataSource {
	return &datasource.CampaignDataSource{DataSource: datasource.DataSource{Session: cms.Session.Copy()}}
}

// return instance of sound file  datasource
// every time a new instance would be created
func (cms *CampaignService) soundFileDatasource() *datasource.SoundFileInfoDataSource {
	return &datasource.SoundFileInfoDataSource{DataSource: datasource.DataSource{Session: cms.Session.Copy()}}
}
