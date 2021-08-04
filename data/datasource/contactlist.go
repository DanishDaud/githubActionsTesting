package datasource

import (
	"errors"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const cContactLists = "contactlists"

// Data Structure to represent ContactListDatasource
type ContactListDataSource struct {
	DataSource
}

func (clds *ContactListDataSource) ContactListWithId(objectId bson.ObjectId) (*model.ContactList, error) {
	contactList := model.ContactList{}
	err := clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).FindId(objectId).One(&contactList)
	if err != nil {
		return nil, errors.New("can not find contact list")
	}

	return &contactList, nil
}


func (clds *ContactListDataSource) TTSListWithId(objectId bson.ObjectId) (*model.TTSList, error) {
	contactList := model.TTSList{}
	err := clds.DbSession().DB(cmlutils.DefaultDatabase()).C("ttslist").FindId(objectId).One(&contactList)
	if err != nil {
		return nil, errors.New("can not find contact list")
	}

	return &contactList, nil
}


func (clds *ContactListDataSource) ContactListWithIdAndUserId(objectId bson.ObjectId, userId bson.ObjectId) (*model.ContactList, error) {
	contactList := model.ContactList{}
	err := clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).Find(bson.M{"_id": objectId, "userid": userId}).One(&contactList)
	if err != nil {
		return nil, errors.New("can not find contact list")
	}

	return &contactList, nil
}
func (cgds *ContactGroupDataSource) DeleteDNCNumber(contactListId bson.ObjectId, number string) error {

	query := bson.M{"contactListId": contactListId}
	update := bson.M{"$pull": bson.M{"numbers": bson.M{"number": number}}}
	updateInfo, err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).UpdateAll(query, update)
	if err != nil {
		logrus.Errorln(err.Error())
	}
	logrus.Infoln(updateInfo.Updated)
	contactGroup := model.ContactGroup{}
	err1 := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).Find(bson.M{"contactListId": contactListId}).One(&contactGroup)
	if err1 != nil {
		logrus.Errorln(err1.Error())
	}
	tnumber := (len(contactGroup.Numbers))
	totalnumber := int32(tnumber)
	contactGroup.TotalNumbers = totalnumber

	query1 := bson.M{"_id": contactGroup.ID}
	update1 := bson.M{"$set": bson.M{"totalNumbers": totalnumber}}
	if err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).Update(query1, update1); err != nil {
		return err
	}
	contactList := model.ContactList{}
	contactList.NumberCount = int64(totalnumber)
	query2 := bson.M{"_id": contactListId}
	update2 := bson.M{"$set": bson.M{"numberCount": totalnumber}}
	if err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).Update(query2, update2); err != nil {
		return err
	}
	return nil
}
func (clds *ContactListDataSource) DoNotContactListWithUserId(userId bson.ObjectId) (model.ContactLists, error) {
	contactList := model.ContactLists{}
	err := clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).Find(bson.M{"userid": userId, "type": 3}).All(&contactList)
	if err != nil {
		return nil, errors.New("can not find contact list")
	}

	return contactList, nil
}
func (clds *ContactListDataSource) SaveContactList(contactList *model.ContactList) error {
	// if there is no contact list id assign one

	if contactList.ID == "" {
		contactList.ID = bson.NewObjectId()
	}

	existingContactListObject, err := clds.ContactListWithId(contactList.ID)
	if err != nil {
		// contact list doesn't exist, create new
		contactList.CreateDate = time.Now().UTC()
		contactList.UpdateDate = time.Now().UTC()
		// Write the contact list to mongo
		if err := clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).Insert(contactList); err != nil {
			return err
		}
	} else {
		// contact list exists
		contactList.UpdateDate = time.Now().UTC()
		if err := clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).UpdateId(existingContactListObject.ID, contactList); err != nil {
			return err
		}
	}
	return nil
}

func (clds *ContactListDataSource) DeleteContactList(contactList *model.ContactList) error {
	// delete contactlist file object from database
	if err := clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).RemoveId(contactList.ID); err != nil {
		// contact list object delete fail
		logrus.Errorln("Failed to Delete Contact List", err.Error(), contactList.ID)
		return err
	}
	return nil
}

func (clds *ContactListDataSource) ContactListGetList(userObjectId bson.ObjectId, page int, limit int, ctype int8, all bool) (*model.ContactLists, int, error) {

	contactListObjects := model.ContactLists{}

	filter := bson.M{"userid": userObjectId, "type": ctype}

	if !all {
		filter["status"] = model.ContactListStatusActive
	}
	query := clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).Find(filter)
	query1 := clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).Find(filter)

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

func (clds *ContactListDataSource) ContactListGetListWithDNCScrubStatus(status model.DNCScrubProcessStatus, page int, limit int) (*model.ContactLists, int, error) {

	contactListObjects := model.ContactLists{}

	filter := bson.M{

		"dncScrubStatus": status,
		"status":         model.ContactListStatusProcessing,
	}

	query := clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).Find(filter)
	query1 := clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).Find(filter)

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

// this method updated total numbers count in dnc list
func (clds *ContactListDataSource) UpdateDNCTotalNumberCount(userID bson.ObjectId, totalNumber int64) error {
	return clds.updateTotalNumberCountContactLists(userID, totalNumber, true)
}

// private method
func (clds *ContactListDataSource) updateTotalNumberCountContactLists(userId bson.ObjectId, totalNumber int64, isDNC bool) error {
	query := bson.M{"userid": userId, "isDoNotContactList": isDNC}
	update := bson.M{"$set": bson.M{"numberCount": totalNumber}}
	return clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).Update(query, update)
}

func (clds *ContactListDataSource) UpdateDNCScrubStatus(id bson.ObjectId, status model.DNCScrubProcessStatus) error {
	query := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"dncScrubStatus": status}}
	return clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).Update(query, update)
}

func (clds *ContactListDataSource) UpdateStatus(id bson.ObjectId, status model.ContactListStatus) error {
	query := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": status}}
	return clds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).Update(query, update)
}

func (cgds *ContactListDataSource) IncrementContactListNumberCount(cgId bson.ObjectId) error {
	query := bson.M{"_id": cgId}
	update := bson.M{"$inc": bson.M{"numberCount": 1}}
	return cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).Update(query, update)
}

func (cgds *ContactListDataSource) DecrementContactListNumberCount(cgId bson.ObjectId) error {
	query := bson.M{"_id": cgId}
	update := bson.M{"$inc": bson.M{"numberCount": -1}}
	return cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactLists).Update(query, update)
}
