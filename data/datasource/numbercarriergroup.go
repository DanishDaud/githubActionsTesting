package datasource

import (
	"time"

	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const ccollection = "numbercarriergroup"

// Data Structure to represent Campaign
type TelnyxNumbersGroupDataSource struct {
	DataSource
}

func (tnd *TelnyxNumbersGroupDataSource) Store(object *model.TelnyxNumbersGroup) (*model.TelnyxNumbersGroup, error) {
	// if there is no campaign id assign one
	if object.ID == "" {
		object.ID = bson.NewObjectId()
		object.CreateDate = time.Now().UTC()
		object.UpdateDate = time.Now().UTC()
		// Write the campaign to mongo
		return object, tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(ccollection).Insert(object)

	} else {
		object.UpdateDate = time.Now().UTC()
		// Write the campaign to mongo
		return object, tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(ccollection).UpdateId(object.ID, object)
	}
}

func (tnd *TelnyxNumbersGroupDataSource) Delete(object *model.TelnyxNumbersGroup) error {
	// if there is no campaign id assign one
	var entry model.TelnyxNumbersGroup
	err := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(ccollection).Find(bson.M{
		"numberID": object.NumberID, "contactListID": object.ContactListID, "userID": object.UserID}).One(&entry)
	if err != nil {
		return err
	}
	return tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(ccollection).Remove(bson.M{"_id": entry.ID})

}

func (tnd *TelnyxNumbersGroupDataSource) GetByID(object *model.TelnyxNumbersGroup) (*model.TelnyxNumbersGroup, error) {
	// if there is no campaign id assign one
	var entry model.TelnyxNumbersGroup

	err := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(ccollection).Find(bson.M{
		"numberID": object.NumberID, "contactListID": object.ContactListID, "userID": object.UserID}).One(&entry)
	if err != nil {
		return nil, err
	}
	return &entry, nil

}

func (tnd *TelnyxNumbersGroupDataSource) GetAllCount(pipeline []bson.M) (int, error) {

	var result []bson.M
	var count int
	stagecount := bson.M{"$count": "total_count"}
	pipeline = append(pipeline, stagecount)
	err1 := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(ccollection).Pipe(pipeline).All(&result)
	if err1 != nil {
		return 0, err1
	}
	if result == nil {
		count = 0
	} else {
		count = (result[0]["total_count"]).(int)
	}
	return count, nil
}

func (tnd *TelnyxNumbersGroupDataSource) fetchBasePipeline(userObjectId bson.ObjectId) ([]bson.M, error) {

	match := bson.M{
		"userID": userObjectId,
	}
	lookup1 := bson.M{
		"from":         "numbers",
		"localField":   "numberID",
		"foreignField": "_id",
		"as":           "numbers",
	}

	lookup2 := bson.M{
		"from":         "contactlists",
		"localField":   "contactListID",
		"foreignField": "_id",
		"as":           "contactList",
	}

	pipeline := []bson.M{
		{
			"$lookup": lookup1,
		},
		{
			"$lookup": lookup2,
		},
		{
			"$match": match,
		},
	}
	return pipeline, nil

}

func (tnd *TelnyxNumbersGroupDataSource) fetchContactListPipeline(userObjectId bson.ObjectId, contactid bson.ObjectId) ([]bson.M, error) {

	var pipeline []bson.M
	match := bson.M{"$match": bson.M{"userID": userObjectId}}

	if contactid != "" {
		match2 := bson.M{"$match": bson.M{"contactListID": contactid}}
		pipeline = append(pipeline, match2)
	}

	group := bson.M{
		"$group": bson.M{"_id": "$contactListID",
			"totalInbound":    bson.M{"$sum": "$totalInbound"},
			"totalOutbound":   bson.M{"$sum": "$totalOutbound"},
			"inboundDuration": bson.M{"$sum": "$inboundDuration"},
		},
	}

	pipeline = append(pipeline, match)
	pipeline = append(pipeline, group)

	//
	//pipeline := []bson.M{
	//	{
	//		"$match": match,
	//
	//	},
	//
	//	{
	//		"$group": group,
	//	},
	//
	//}

	return pipeline, nil

}

func (tnd *TelnyxNumbersGroupDataSource) GetOneTelnyxNumberGroup(object model.TelnyxNumbersGroup) (*model.TelnyxNumbersGroup, error) {
	// if there is no campaign id assign one
	var entry model.TelnyxNumbersGroup

	err := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(ccollection).Find(bson.M{
		"numberID": object.NumberID, "contactListID": object.ContactListID, "userID": object.UserID, "contactGroupID": object.ContactGroupID}).One(&entry)
	if err != nil {
		return nil, err
	}
	return &entry, nil

}

func (tnd *TelnyxNumbersGroupDataSource) DeleteNumberGroupsByContactListID(userid bson.ObjectId, cid bson.ObjectId) error {

	_, err := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(ccollection).RemoveAll(bson.M{"userID": userid, "contactListID": cid})
	if err != nil {
		return err
	}
	return nil

}

func (tnd *TelnyxNumbersGroupDataSource) GetAll(op model.GetAllTelnyxGroupRequest) (*[]model.TelnyxNumbersGroupAggr, int, error) {

	regex := bson.M{"$regex": op.Search, "$options": "i"}
	search := bson.M{"$match": bson.M{"numbers.phoneNumber": regex}}
	sort := bson.M{"$sort": bson.M{"createDate": -1}}
	skip := bson.M{"$skip": ((op.Page - 1) * op.Limit)}
	limitt := bson.M{"$limit": op.Limit}
	match := bson.M{"$match": bson.M{"numberID": op.NumberID}}
	var aggrobj []model.TelnyxNumbersGroupAggr
	if op.Page < 1 || op.Limit < 1 {
		pipeline, _ := tnd.fetchBasePipeline(op.UserID)
		if op.NumberID != "" {
			pipeline = append(pipeline, match)
		}
		pipe := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(ccollection).Pipe(pipeline)
		if err := pipe.All(&aggrobj); err != nil {
			return nil, 0, err
		}
		count, _ := tnd.GetAllCount(pipeline)
		return &aggrobj, count, nil

	} else {
		pipeline, _ := tnd.fetchBasePipeline(op.UserID)
		if op.NumberID != "" {
			pipeline = append(pipeline, match)
		}
		pipeline = append(pipeline, sort)
		pipeline = append(pipeline, skip)
		pipeline = append(pipeline, limitt)
		pipeline = append(pipeline, sort)
		pipeline = append(pipeline, search)

		pipe := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(ccollection).Pipe(pipeline)
		if err := pipe.All(&aggrobj); err != nil {
			return nil, 0, err
		}

		count, _ := tnd.GetAllCount(pipeline)
		return &aggrobj, count, nil
	}

}

func (tnd *TelnyxNumbersGroupDataSource) GetContactListStats(op model.GetContactListStatsRequest) (*[]model.ContactListStats, int, error) {

	//regex :=  bson.M{"$regex": op.Search, "$options": "i",}
	//search := bson.M{ "$match": bson.M{"contactListID" : regex}}
	//sort := bson.M{"$sort": bson.M{"createDate":-1}}
	skip := bson.M{"$skip": ((op.Page - 1) * op.Limit)}
	limitt := bson.M{"$limit": op.Limit}
	match := bson.M{"$match": bson.M{"contactListID": op.ContactListID}}
	var aggrobj []model.ContactListStats
	if op.Page < 1 || op.Limit < 1 {
		pipeline, _ := tnd.fetchContactListPipeline(op.UserID, op.ContactListID)
		if op.ContactListID != "" {
			pipeline = append(pipeline, match)
		}
		pipe := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(ccollection).Pipe(pipeline)
		if err := pipe.All(&aggrobj); err != nil {
			return nil, 0, err
		}
		count, _ := tnd.GetAllCount(pipeline)
		return &aggrobj, count, nil

	} else {

		pipeline, _ := tnd.fetchContactListPipeline(op.UserID, op.ContactListID)
		//pipeline=append(pipeline,sort)
		pipeline = append(pipeline, skip)
		pipeline = append(pipeline, limitt)
		//pipeline=append(pipeline,sort)
		//pipeline=append(pipeline,search)

		pipe := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(ccollection).Pipe(pipeline)
		if err := pipe.All(&aggrobj); err != nil {
			return nil, 0, err
		}

		count, _ := tnd.GetAllCount(pipeline)
		return &aggrobj, count, nil
	}

}

//func (tnd *TelnyxDataSource) Save(object *model.TelnyxNumber) error {
//	// if there is no campaign id assign one
//	if object.ID == "" {
//		object.ID = bson.NewObjectId()
//		object.CreateDate = time.Now().UTC()
//		object.UpdateDate = time.Now().UTC()
//		// Write the campaign to mongo
//		return tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Insert(object)
//	} else {
//		object.UpdateDate = time.Now().UTC()
//		// Write the campaign to mongo
//		return tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).UpdateId(object.ID, object)
//	}
//}
//
//func (tnd *TelnyxDataSource) DeleteWithId(id string) error {
//	return tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Remove(bson.M{"phone_number": id})
//}
//
//
//func (tnd *TelnyxDataSource) Update(object *model.TelnyxNumber) error {
//	colQuerier := bson.M{"phoneNumber": object.PhoneNumber}
//	change := bson.M{"$set": bson.M{"phoneStatus": object.PhoneStatus}}
//	return tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Update(colQuerier,change)
//}
//
//func (tnd *TelnyxDataSource) Get(object string) (model.TelnyxNumber,  error) {
//	var contactListObjects model.TelnyxNumber
//	_ = tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Find(bson.M{"phoneNumber": object}).One(&contactListObjects)
//	return contactListObjects,nil
//}
//
//
//
//func (tnd *TelnyxDataSource) TelnyxListGetList(userObjectId bson.ObjectId, page int, limit int) (*model.TelnyxList, int, error) {
//
//	contactListObjects := model.TelnyxList{}
//
//	query := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Find(bson.M{"userid": userObjectId})
//	query1 := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Find(bson.M{"userid": userObjectId})
//	fmt.Println(query.All(&contactListObjects))
//	fmt.Println(query)
//	if page < 1 || limit < 1 {
//		if err := query.All(&contactListObjects); err != nil {
//			fmt.Println(err)
//			return nil, 0, err
//		}
//	} else {
//		if err := query.Sort("-createDate").Skip((page - 1) * limit).Limit(limit).All(&contactListObjects); err != nil {
//			return nil, 0, err
//		}
//	}
//
//	count, er := query1.Count()
//	if er != nil {
//		return nil, 0, er
//	}
//
//	return &contactListObjects, count, nil
//}
