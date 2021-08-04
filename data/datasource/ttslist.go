package datasource

import (
	"errors"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
	"time"
)

const cttsfile = "ttsfile"
const cttslist = "ttslist"

type TTSListDataSource struct {
	DataSource
}




func (uds *TTSListDataSource) SaveFile(obj model.TTSFile) (*model.TTSFile, error) {

	if obj.ID == "" {
		obj.ID = bson.NewObjectId()
	}

	err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cttsfile).Insert(obj)

	if err != nil {
		return nil, err
	}

	return &obj, nil

}

func (uds *TTSListDataSource) GET(userid bson.ObjectId, listid bson.ObjectId) (*model.TTSList, error) {

	var result model.TTSList
	err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cttslist).Find(bson.M{"userid": userid, "_id": listid}).One(&result)

	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (uds *TTSListDataSource) Delete(listid bson.ObjectId) error {

	err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cttslist).Remove(bson.M{"_id": listid})

	if err != nil {
		return err
	}
	return nil
}

func (uds *TTSListDataSource) List(obj model.CustomFilter, all bool) (*[]model.TTSList, int, error) {

	if obj.Page < 1 {
		obj.Page = 1
	}
	if obj.Limit < 1 {
		obj.Limit = 10
	}
	var result []model.TTSList
	filter := bson.M{"userid": obj.UserId}

	if !all {
		filter["status"] = model.ContactListStatusActive
	}

	query := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cttslist).Find(filter)
	query1 := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cttslist).Find(filter)
	c, err := query.Count()
	if err != nil {
		return nil, 0, err
	}
	query1.Sort("-createDate").Skip((obj.Page - 1) * obj.Limit).Limit(obj.Limit).All(&result)
	if err != nil {
		return nil, 0, err
	}
	return &result, c, nil

}

func (uds *TTSListDataSource) SaveTTSList(obj model.TTSList) error {
	// if there is no contact list id assign one
	if obj.ID == "" {
		obj.ID = bson.NewObjectId()
	}
	existingContactListObject, err := uds.TTSListWithId(obj.ID)
	if err != nil {
		// contact list doesn't exist, create new
		obj.CreateDate = time.Now().UTC()
		obj.UpdateDate = time.Now().UTC()
		// Write the contact list to mongo
		if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cttslist).Insert(obj); err != nil {
			return err
		}
	} else {
		// contact list exists
		obj.UpdateDate = time.Now().UTC()
		if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cttslist).UpdateId(existingContactListObject.ID, obj); err != nil {
			return err
		}
	}
	return nil
}
func (uds *TTSListDataSource) GetFile(objectId bson.ObjectId) (*model.TTSFile, error) {
	result := model.TTSFile{}
	err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cttsfile).Find(bson.M{"_id": objectId}).One(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil

}

func (uds *TTSListDataSource) TTSListWithId(objectId bson.ObjectId) (*model.TTSList, error) {
	contactList := model.TTSList{}
	err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cttslist).FindId(objectId).One(&contactList)
	if err != nil {
		return nil, errors.New("can not find contact list")
	}

	return &contactList, nil
}




func (dnc *TTSListDataSource) SaveDncJob(dncjobs *model.DNCJobs) error {
	// if there is no contact list id assign one

	if dncjobs.ID == "" {
		dncjobs.ID = bson.NewObjectId()
	}

	existingdncJobObject, err := dnc.DncJobWithId(dncjobs.ID)
	if err != nil {
		// contact list doesn't exist, create new
		dncjobs.CreateDate = time.Now().UTC()
		dncjobs.UpdateDate = time.Now().UTC()
		// Write the contact list to mongo
		if err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobs).Insert(dncjobs); err != nil {
			return err
		}
	} else {
		// contact list exists
		dncjobs.UpdateDate = time.Now().UTC()
		if err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobs).UpdateId(existingdncJobObject.ID, dncjobs); err != nil {
			return err
		}
	}
	return nil
}


func (dnc *TTSListDataSource) DncJobWithId(objectId bson.ObjectId) (*model.DNCJobs, error) {
	dncjobObject := model.DNCJobs{}
	err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobs).FindId(objectId).One(&dncjobObject)
	if err != nil {
		return nil, errors.New("can not find dnc job object")
	}

	return &dncjobObject, nil
}

type TTSLists []model.TTSList

func (clds *TTSListDataSource) GetListsWithDNCScrubStatus(status model.DNCScrubProcessStatus, page int, limit int) (*TTSLists, int, error) {

	contactListObjects := TTSLists{}

	filter := bson.M{
		"dncScrubStatus": status,
		"status":         model.ContactListStatusProcessing,
	}

	query := clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cttslist).Find(filter)
	query1 := clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cttslist).Find(filter)

	if page < 1 || limit < 1 {
		if err := query.Sort("-createDate").All(&contactListObjects); err != nil {
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

func (clds *TTSListDataSource) UpdateStatus(id bson.ObjectId, status model.ContactListStatus) error {
	query := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": status}}
	return clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cttslist).Update(query, update)
}


func (clds *TTSListDataSource) UpdateDNCScrubStatus(id bson.ObjectId, status model.DNCScrubProcessStatus) error {
	query := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"dncScrubStatus": status}}
	return clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cttslist).Update(query, update)
}
