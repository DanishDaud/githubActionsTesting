package datasource

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
	"time"
)

const dncJobs = "dncjobs"

// Data Structure to represent Campaign
type DNCJobsDataSource struct {
	DataSource
}

func (dnc *DNCJobsDataSource) SaveDNCJob(dncjobs *model.DNCJobs) error {
	// if there is no contact list id assign one

	if dncjobs.ID == "" {
		dncjobs.ID = bson.NewObjectId()
	}

	existingdncJobObject, err := dnc.DNCJobWithId(dncjobs.ID)
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

func (dnc *DNCJobsDataSource) DNCJobWithId(objectId bson.ObjectId) (*model.DNCJobs, error) {
	dncjobObject := model.DNCJobs{}
	err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobs).FindId(objectId).One(&dncjobObject)
	if err != nil {
		return nil, errors.New("can not find dnc job object")
	}

	return &dncjobObject, nil
}

func (dnc *DNCJobsDataSource) DeleteDncJOB(dncobj *model.DNCJobs) error {
	// delete contactlist file object from database
	if err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobs).RemoveId(dncobj.ID); err != nil {
		// contact list object delete fail
		logrus.Errorln("Failed to Delete dnc job object", err.Error(), dncobj.ID)
		return err
	}
	return nil
}

func (dnc *DNCJobsDataSource) DNCObjectWithContactlistId(objectId bson.ObjectId, page int, limit int) (*model.DNCJobsList, int, error) {
	// delete contactlist file object from database
	dncjobObjects := model.DNCJobsList{}

	filter := bson.M{"contactListId": objectId}

	query := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobs).Find(filter)
	query1 := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobs).Find(filter)
	if page < 1 || limit < 1 {
		if err := query.Sort("-createDate").All(&dncjobObjects); err != nil {
			return nil, 0, err
		}
	} else {
		if err := query.Sort("-createDate").Skip((page - 1) * limit).Limit(limit).All(&dncjobObjects); err != nil {
			return nil, 0, err
		}
	}

	count, er := query1.Count()
	if er != nil {
		return nil, 0, er
	}

	return &dncjobObjects, count, nil
}

func (dnc *DNCJobsDataSource) DNCObjectWithJobId(objectId string) (*model.DNCJobs, error) {
	dncjobList := model.DNCJobs{}
	err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobs).Find(bson.M{"jobid": objectId}).One(&dncjobList)
	if err != nil {
		return nil, errors.New("can not find dnc job object")
	}

	return &dncjobList, nil
}

func (dnc *DNCJobsDataSource) DNCJobCountWithContactListId(contactlistId bson.ObjectId) (int, error) {
	c, err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobs).Find(bson.M{"contactListId": contactlistId}).Count()
	if err != nil {
		// contact list object delete fail
		logrus.Errorln("Failed to get  dnc job count", err.Error(), contactlistId.Hex())
		return 0, err
	}
	return c, nil
}

func (dnc *DNCJobsDataSource) DNCJobCountWithTTSId(ttsListId bson.ObjectId) (int, error) {
	c, err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobs).Find(bson.M{"ttsListID": ttsListId}).Count()
	if err != nil {
		// contact list object delete fail
		logrus.Errorln("Failed to get  dnc job count", err.Error(), ttsListId.Hex())
		return 0, err
	}
	return c, nil
}

func (dnc *DNCJobsDataSource) DNCObjectsWithStatus(status model.DNCJobType) (*model.DNCJobsList, error) {
	dncjobList := model.DNCJobsList{}
	err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobs).Find(bson.M{"status": status}).All(&dncjobList)
	if err != nil {
		return nil, errors.New("can not find dnc job object")
	}

	return &dncjobList, nil
}
