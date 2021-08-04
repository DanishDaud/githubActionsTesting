package datasource

import (
	"errors"
	"strings"
	"time"

	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const cCdr = "cdr"
const uDispositionStats = "userdispositionstats"
const alegStatshourly = "userreportaleghourly"
const alegStats = "userreportaleg"

const blegStats = "userreportbleg"

// Data Structure to represent Campaign
type CdrDataSource struct {
	DataSource
}

func (cdrds *CdrDataSource) GetById(objectId bson.ObjectId) (*model.CDR, error) {
	cdr := model.CDR{}
	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).FindId(objectId).One(&cdr)
	if err != nil {
		return nil, errors.New("can not find cdr")
	}

	return &cdr, nil
}

func (cdrds *CdrDataSource) Save(obj *model.CDR) error {
	// if there is no campaign id assign one
	if obj.ID == "" {
		obj.ID = bson.NewObjectId()
	}

	existingCampaignObject, err := cdrds.GetById(obj.ID)
	if err != nil {
		// campaign doesn't exist, create new
		obj.CreateDate = time.Now().UTC()
		obj.UpdateDate = time.Now().UTC()
		// Write the campaign to mongo
		if err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Insert(obj); err != nil {
			return err
		}
	} else {
		// cdr exists
		obj.UpdateDate = time.Now().UTC()
		if err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).UpdateId(existingCampaignObject.ID, obj); err != nil {
			return err
		}
	}
	return nil
}

func (cdrds *CdrDataSource) CdrListForCampaigns(cids []bson.ObjectId, page int, limit int) (*model.CampaingList, int, error) {
	list := model.CampaingList{}

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	query := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Find(bson.M{"_id": bson.M{"$in": cids}})
	query1 := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Find(bson.M{"_id": bson.M{"$in": cids}})
	if err := query.Sort("-createDate").Skip((page - 1) * limit).Limit(limit).All(&list); err != nil {
		return nil, 0, err
	}

	count, er := query1.Count()
	if er != nil {
		return nil, 0, er
	}

	return &list, count, nil
}

func (cdrds *CdrDataSource) CdrList(userObjectId bson.ObjectId, page int, limit int, from string, to string, amdStatus int8, callStatus int8, callType int, campaign string) (model.CdrList, int, error) {
	var andQuery []map[string]interface{}
	list := model.CdrList{}
	var RFC3339 = "2006-01-02T15:04:05Z07:00"
	var fromDate time.Time
	var toDate time.Time

	if to == "" && from == "" && amdStatus == 0 && callStatus == 0 && callType == 0 && campaign == "" {
		query := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Find(bson.M{"userid": userObjectId})
		query1 := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Find(bson.M{"userid": userObjectId})
		if page < 1 {
			page = 1
		}

		if limit < 1 {
			limit = 10
		}

		if err := query.Sort("-$natural").Skip((page - 1) * limit).Limit(limit).All(&list); err != nil {
			return nil, 0, err
		}

		count, er := query1.Count()
		if er != nil {
			return nil, 0, er
		}

		return list, count, nil
	} else {

		if from != "" {

			s := strings.SplitAfter(from, "Z")
			from = (s[0])
			fromDate, _ = time.Parse(RFC3339, from)

			q0 := bson.M{"startDate": bson.M{"$gte": fromDate}}
			andQuery = append(andQuery, q0)
		}

	}
	if to != "" {
		s := strings.SplitAfter(to, "Z")
		to = (s[0])
		toDate, _ = time.Parse(RFC3339, to)

		q1 := bson.M{"startDate": bson.M{"$lte": toDate}}
		andQuery = append(andQuery, q1)
	}
	if amdStatus != 0 {
		q2 := bson.M{"amdStatus": amdStatus}
		andQuery = append(andQuery, q2)
	}
	if callStatus != 0 {
		q3 := bson.M{"disposition": callStatus}
		andQuery = append(andQuery, q3)
	}
	if callType != 0 {
		q4 := bson.M{"calltype": callType}
		andQuery = append(andQuery, q4)
	}
	if campaign != "" {
		id := bson.ObjectIdHex(campaign)
		q5 := bson.M{"campaignId": id}
		andQuery = append(andQuery, q5)
	}
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	query := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Find(bson.M{"userid": userObjectId, "$and": andQuery})
	query1 := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Find(bson.M{"userid": userObjectId, "$and": andQuery})

	if err := query.Sort("-$natural").Skip((page - 1) * limit).Limit(limit).All(&list); err != nil {
		return nil, 0, err
	}

	count, er := query1.Count()
	if er != nil {
		return nil, 0, er
	}

	return list, count, nil
}
