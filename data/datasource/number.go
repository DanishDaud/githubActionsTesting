package datasource

import (
	"time"

	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const collection = "numbers"
const collection2 = "numbercarriergroup"
const collection3 = "contactgroups"

// Data Structure to represent Campaign
type NumberDataSource struct {
	DataSource
}

func (tnd *NumberDataSource) Store(object *model.TelnyxNumber) (*model.TelnyxNumber, error) {
	// if there is no campaign id assign one
	if object.ID == "" {
		object.ID = bson.NewObjectId()
		object.CreateDate = time.Now().UTC()
		object.UpdateDate = time.Now().UTC()
		// Write the campaign to mongo
		return object, tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Insert(object)
	} else {
		object.UpdateDate = time.Now().UTC()
		// Write the campaign to mongo
		return object, tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).UpdateId(object.ID, object)
	}
}

func (tnd *NumberDataSource) DeleteById(id bson.ObjectId) error {
	return tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Remove(bson.M{"_id": id})
}

func (tnd *NumberDataSource) DeleteTelnyxGroupByID(id bson.ObjectId) error {
	return tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection2).Remove(bson.M{"_id": id})
}

func (tnd *NumberDataSource) DeleteTelnyxNumberFromContactGroup(contactgroupid bson.ObjectId, number string) error {
	query := bson.M{"_id": contactgroupid}
	update := bson.M{"$pull": bson.M{"numbers": bson.M{"number": number}}}
	if err := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection3).Update(query, update); err != nil {
		// contact group delete fail
		return err
	}
	return nil
}

func (tnd *NumberDataSource) GetNumberCarrierGroupByNumberID(id bson.ObjectId) ([]model.TelnyxNumbersGroup, error) {
	var contactListObjects []model.TelnyxNumbersGroup
	err := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection2).Find(bson.M{"numberID": id}).All(&contactListObjects)
	if err != nil {
		return contactListObjects, err
	}
	return contactListObjects, nil
}

func (tnd *NumberDataSource) UpdateByStatus(object *model.TelnyxNumber) (*model.TelnyxNumber, error) {
	colQuerier := bson.M{"_id": object.ID}
	change := bson.M{"$set": bson.M{"phoneStatus": object.PhoneStatus, "updateDate": time.Now().UTC()}}
	err := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Update(colQuerier, change)
	if err != nil {
		return nil, err
	}
	result, _ := tnd.GetByID(object.ID)
	return result, err
}

func (tnd *NumberDataSource) UpdateByForwardingNumber(object *model.TelnyxNumber) (*model.TelnyxNumber, error) {
	colQuerier := bson.M{"_id": object.ID}
	change := bson.M{"$set": bson.M{"forwardingNumber": object.ForwardingNumber, "updateDate": time.Now().UTC()}}
	err := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Update(colQuerier, change)
	if err != nil {
		return nil, err
	}
	result, _ := tnd.GetByID(object.ID)
	return result, err

}

func (tnd *NumberDataSource) GetByNumber(object string) (*model.TelnyxNumber, error) {
	var contactListObjects model.TelnyxNumber
	err := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Find(bson.M{"phoneNumber": object}).One(&contactListObjects)
	if err != nil {
		return nil, err
	}
	return &contactListObjects, nil
}

func (tnd *NumberDataSource) GetByID(id bson.ObjectId) (*model.TelnyxNumber, error) {
	var contactListObjects model.TelnyxNumber
	err := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Find(bson.M{"_id": id}).One(&contactListObjects)
	if err != nil {
		return nil, err
	}
	return &contactListObjects, nil
}

func (tnd *NumberDataSource) GetForwardingNumberByNumber(number string) (string, error) {
	var contactListObjects model.TelnyxNumber
	err := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Find(bson.M{"phoneNumber": number}).One(&contactListObjects)
	if err != nil {
		return "", err
	}
	return contactListObjects.ForwardingNumber, nil
}

func (tnd *NumberDataSource) GetAll(op model.GetAllRequest) (*model.TelnyxList, int, error) {
	contactListObjects := model.TelnyxList{}
	regex := op.Search
	_ = tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Find(bson.M{"userID": op.UserID})
	query1 := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Find(bson.M{"userID": op.UserID, "phoneNumber": bson.M{"$regex": regex}})
	//fmt.Println(query.All(&contactListObjects))
	//fmt.Println(query)
	if op.Page < 1 || op.Limit < 1 {
		if err := query1.All(&contactListObjects); err != nil {

			return nil, 0, err
		}
	} else {
		query1 = query1.Sort("-createDate").Skip((op.Page - 1) * op.Limit).Limit(op.Limit)
		err := query1.All(&contactListObjects)
		if err != nil {
			return nil, 0, err
		}
	}
	count, er := query1.Count()
	if er != nil {
		return nil, 0, er
	}
	return &contactListObjects, count, nil
}

//func (tnd *TelnyxDataSource) UpdateByNumber(num string,number string) error {
//	colQuerier := bson.M{"phoneNumber": num}
//	change := bson.M{"$set": bson.M{"forwardingNumber": number,"updateDate": time.Now().UTC() }}
//	return tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Update(colQuerier,change)
//}
//
//func (tnd *TelnyxDataSource) DeleteByNumber(num string) error {
//	return tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collection).Remove(bson.M{"phoneNumber": num})
//}
//

//func (tnd *TelnyxDataSource) SaveTelnyxCID(object *model.TelnyxNumberCID) error {
//	// if there is no campaign id assign one
//	if object.ID == "" {
//		object.ID = bson.NewObjectId()
//		object.CreateDate = time.Now().UTC()
//		object.UpdateDate = time.Now().UTC()
//		// Write the campaign to mongo
//		return tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collTelCid).Insert(object)
//	} else {
//		object.UpdateDate = time.Now().UTC()
//		// Write the campaign to mongo
//		return tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(collTelCid).UpdateId(object.ID, object)
//	}
//}
