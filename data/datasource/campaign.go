package datasource

import (
	"errors"
	"time"

	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlconstants"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const cCampaigns = "campaigns"

// Data Structure to represent Campaign
type CampaignDataSource struct {
	DataSource
}

func (cds *CampaignDataSource) RunningCampaignsCountForUser(userId bson.ObjectId) (int, error) {
	query := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Find(bson.M{"userid": userId, "status": cmlconstants.CampaignStatusRunning})
	return query.Count()
}

func (cds *CampaignDataSource) CountWithLiveSoundFileId(userId bson.ObjectId, soundFileId bson.ObjectId) (int, error) {
	query := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Find(bson.M{"userid": userId, "call.soundFileId": soundFileId})
	return query.Count()
}

func (cds *CampaignDataSource) CountWithVMSoundFileId(userId bson.ObjectId, soundFileId bson.ObjectId) (int, error) {
	query := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Find(bson.M{"userid": userId, "call.vmSettings.vmSoundFileId": soundFileId})
	return query.Count()
}

func (cds *CampaignDataSource) CountWithTransferSoundFileId(userId bson.ObjectId, soundFileId bson.ObjectId) (int, error) {
	query := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Find(bson.M{"userid": userId, "call.transfer.soundFileId": soundFileId})
	return query.Count()
}

func (cds *CampaignDataSource) CountWithDNCSoundFileId(userId bson.ObjectId, soundFileId bson.ObjectId) (int, error) {
	query := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Find(bson.M{"userid": userId, "call.dnc.soundFileId": soundFileId})
	return query.Count()
}

func (cds *CampaignDataSource) CountWithContactList(userId bson.ObjectId, objectId bson.ObjectId) (int, error) {
	query := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Find(bson.M{"userid": userId, "call.contactlistId": objectId})
	return query.Count()
}

func (cds *CampaignDataSource) CountWithTTSList(userId bson.ObjectId, objectId bson.ObjectId) (int, error) {
	query := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Find(bson.M{"userid": userId, "call.ttslistId": objectId})
	return query.Count()
}
func (cds *CampaignDataSource) CampaignWithId(objectId bson.ObjectId) (*model.Campaign, error) {
	campaignObject := model.Campaign{}
	err := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).FindId(objectId).One(&campaignObject)
	if err != nil {
		return nil, errors.New("can not find campaign")
	}

	return &campaignObject, nil
}

func (cds *CampaignDataSource) DeleteWithId(objectId bson.ObjectId) error {
	return cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).RemoveId(objectId)
}

func (cds *CampaignDataSource) UpdateName(objectId bson.ObjectId, name string) error {
	return cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Update(bson.M{"_id": objectId}, bson.M{"$set": bson.M{"name": name}})
}

func (cds *CampaignDataSource) UpdateContactListinfo(objectId bson.ObjectId, clId bson.ObjectId, total int64) error {
	return cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Update(bson.M{"_id": objectId}, bson.M{"$set": bson.M{"call.stats.total": total, "contactlistId": clId}})
}

func (cds *CampaignDataSource) UpdateLimits(objectId bson.ObjectId, cpm int, transfer int) error {
	return cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Update(bson.M{"_id": objectId}, bson.M{"$set": bson.M{"call.settings.cpm": cpm, "call.settings.transfer": transfer}})
}

func (cds *CampaignDataSource) CampaignWithCampaignIdAndUserId(objectId bson.ObjectId, userId bson.ObjectId) (*model.Campaign, error) {
	campaignObject := model.Campaign{}
	err := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Find(bson.M{"_id": objectId, "userid": userId}).One(&campaignObject)
	if err != nil {

		return nil, errors.New("can not find campaign")
	}

	return &campaignObject, nil
}

func (cds *CampaignDataSource) CampaignListGetList(userObjectId bson.ObjectId, page int, limit int) (*model.CampaingList, int, error) {

	contactListObjects := model.CampaingList{}

	query := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Find(bson.M{"userid": userObjectId})
	query1 := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Find(bson.M{"userid": userObjectId})

	if page < 1 || limit < 1 {
		if err := query.All(&contactListObjects); err != nil {
			return nil, 0, err
		}
	} else {
		if err := query.Sort("-createDate").Skip((page - 1) * limit).Limit(limit).All(&contactListObjects); err != nil {
			return nil, 0, err
		}
	}

	count, er := query1.Count()
	if er != nil {
		return nil, 0, er
	}

	return &contactListObjects, count, nil
}

func (cds *CampaignDataSource) CampaignWithContactId(objectId bson.ObjectId) (model.CampaingList, error) {
	campaignObject := model.CampaingList{}
	err := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Find(bson.M{"call.dncListId": objectId}).All(&campaignObject)
	if err != nil {
		return nil, errors.New("can not find campaign")
	}

	return campaignObject, nil
}

func (cds *CampaignDataSource) SaveCampaign(campaignObject *model.Campaign) error {
	// if there is no campaign id assign one
	if campaignObject.ID == "" {
		campaignObject.ID = bson.NewObjectId()
	}

	existingCampaignObject, err := cds.CampaignWithId(campaignObject.ID)
	if err != nil {
		// campaign doesn't exist, create new
		campaignObject.CreateDate = time.Now().UTC()
		campaignObject.UpdateDate = time.Now().UTC()
		// Write the campaign to mongo
		if err := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Insert(campaignObject); err != nil {
			return err
		}
	} else {
		campaignObject.UpdateDate = time.Now().UTC()
		if err := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).UpdateId(existingCampaignObject.ID, campaignObject); err != nil {
			return err
		}
	}
	return nil
}

func (cds *CampaignDataSource) UpdateCampaignStatus(campaign *model.Campaign, status int8) error {
	if err := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Update(bson.M{"_id": campaign.ID}, bson.M{"$set": bson.M{"status": status}}); err != nil {
		return err
	}
	return nil
}

func (cds *CampaignDataSource) UpdateSchedule(campaign *model.Campaign, id string) error {
	if err := cds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Update(bson.M{"_id": campaign.ID}, bson.M{"$set": bson.M{"call.scheduleSettings.id": id}}); err != nil {
		return err
	}
	return nil
}
