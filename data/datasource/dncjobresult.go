package datasource

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
	"time"
)

const dncJobsResult = "dncjobsresult"

// Data Structure to represent Campaign
type DNCJobsResultDataSource struct {
	DataSource
}

func (dnc *DNCJobsResultDataSource) SaveDncJobResult(dncjobs *model.DNCJobResult) error {
	// if there is no contact list id assign one

	if dncjobs.ID == "" {
		dncjobs.ID = bson.NewObjectId()
	}

	existingdncJobObject, err := dnc.DncJobResultWithId(dncjobs.ID)
	if err != nil {
		// contact list doesn't exist, create new
		dncjobs.CreateDate = time.Now().UTC()
		dncjobs.UpdateDate = time.Now().UTC()
		// Write the contact list to mongo
		if err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobsResult).Insert(dncjobs); err != nil {
			return err
		}
	} else {
		// contact list exists
		dncjobs.UpdateDate = time.Now().UTC()
		if err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobsResult).UpdateId(existingdncJobObject.ID, dncjobs); err != nil {
			return err
		}
	}
	return nil
}

func (dnc *DNCJobsResultDataSource) DncJobResultWithId(objectId bson.ObjectId) (*model.DNCJobResult, error) {
	dncjobObject := model.DNCJobResult{}
	err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobsResult).FindId(objectId).One(&dncjobObject)
	if err != nil {
		return nil, errors.New("can not find contact list")
	}

	return &dncjobObject, nil
}

func (dnc *DNCJobsResultDataSource) DeleteDncJobResult(dncobj *model.DNCJobResult) error {
	// delete contactlist file object from database
	if err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobsResult).RemoveId(dncobj.ID); err != nil {
		// contact list object delete fail
		logrus.Errorln("Failed to Delete Contact List", err.Error(), dncobj.ID)
		return err
	}
	return nil
}

func (dnc *DNCJobsResultDataSource) DncJobResultCountWithContactListId(contactlistId bson.ObjectId) (int, error) {
	c, err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobsResult).Find(bson.M{"contactListId": contactlistId}).Count()
	if err != nil {
		// contact list object delete fail
		logrus.Errorln("Failed to get  dnc job result count", err.Error(), contactlistId.Hex())
		return 0, err
	}
	return c, nil
}




func (dnc *DNCJobsResultDataSource) DNCJobResultCountWithTTSListId(listID bson.ObjectId) (int, error) {
	c, err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobsResult).Find(bson.M{"ttsListID": listID}).Count()
	if err != nil {
		// contact list object delete fail
		logrus.Errorln("Failed to get  dnc job result count", err.Error(), listID.Hex())
		return 0, err
	}
	return c, nil
}


func (dnc *DNCJobsResultDataSource) DncJobResultWithContactListId(contactlistId bson.ObjectId, page int, limit int) (*model.DNCJobsResultList, int, error) {
	// delete contactlist file object from database
	dncjobObjects := model.DNCJobsResultList{}

	filter := bson.M{
		"contactListId": contactlistId,
	}

	query := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobsResult).Find(filter)
	query1 := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobsResult).Find(filter)
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



func (dnc *DNCJobsResultDataSource) DncJobResultWithttsListId(contactlistId bson.ObjectId, page int, limit int) (*model.DNCJobsResultList, int, error) {
	// delete contactlist file object from database
	dncjobObjects := model.DNCJobsResultList{}

	filter := bson.M{
		"ttsListID": contactlistId,
	}

	query := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobsResult).Find(filter)
	query1 := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobsResult).Find(filter)
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


func (dnc *DNCJobsResultDataSource) DncJobResultObjectWithId(objectId bson.ObjectId) (*model.DNCJobResult, error) {
	dncjobList := model.DNCJobResult{}
	err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobsResult).FindId(objectId).One(&dncjobList)
	if err != nil {
		return nil, errors.New("can not find contact list")
	}

	return &dncjobList, nil
}

func (dnc *DNCJobsResultDataSource) DncJobResultObjectWithDncJobId(jobid string) (*model.DNCJobResult, error) {
	dncjobList := model.DNCJobResult{}
	err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobsResult).Find(bson.M{"dnc_job_id": jobid}).One(&dncjobList)
	if err != nil {
		return nil, errors.New("can not find contact list")
	}

	return &dncjobList, nil
}

func (dnc *DNCJobsResultDataSource) DncJobResultObjectWithJobId(jobid string) (*model.DNCJobResult, error) {
	dncjobList := model.DNCJobResult{}
	err := dnc.DbSession().DB(cmlutils.DefaultDatabase()).C(dncJobsResult).Find(bson.M{"jobid": jobid}).One(&dncjobList)
	if err != nil {
		return nil, errors.New("can not find contact list")
	}

	return &dncjobList, nil
}


