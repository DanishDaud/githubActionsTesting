package datasource

import (
	"errors"
	"time"

	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const cContactGroups = "contactgroups"

// Data Structure to represent ContactListDatasource
type ContactGroupDataSource struct {
	DataSource
}
type TimeZone struct {
	Count      string `json:"-" bson:"count"`
	Country    string `json:"-" bson:"countryname"`
	Numbercode string `json:"-" bson:"numbercode"`
	State      string `json:"-" bson:"state"`
	T1         string `json:"-" bson:"t1"`
	T2         string `json:"-" bson:"t2"`
	T3         string `json:"-" bson:"t3"`
	T4         string `json:"-" bson:"t4"`
	TimeZone   string `json:"-" bson:"timezone"`
}
type WirelessNumber struct {
	Numbercode string `json:"-" bson:"wirelessnumber"`
}
type LindlineNumber struct {
	Numbercode string `json:"-" bson:"landlinenumber"`
}

type TelnyxNumberObject struct {
	ID         bson.ObjectId ` bson:"_id"`
	Number     string        ` bson:"number"`
	TimeZone   string        ` bson:"timezone"`
	Numbertype string        ` bson:"numbertype"`
}

func (cgds *ContactGroupDataSource) ContactGroupWithId(objectId bson.ObjectId) (*model.ContactGroup, error) {
	contactGroup := model.ContactGroup{}
	err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).FindId(objectId).One(&contactGroup)
	if err != nil {
		return nil, errors.New("can not find contact group")
	}

	return &contactGroup, nil
}
func (cgds *ContactGroupDataSource) Findtimezone(numbercode string) string {

	timezone := TimeZone{}

	if err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C("timezone").Find(bson.M{"numbercode": numbercode}).One(&timezone); err != nil {
		return ""
	}

	time := timezone.TimeZone

	return time
}
func (cgds *ContactGroupDataSource) FindWirelessNumber(numbercode string) string {

	wirelessnumber := WirelessNumber{}

	if err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C("cellphonenumber").Find(bson.M{"wirelessnumber": numbercode}).One(&wirelessnumber); err != nil {
		return ""
	}

	number := wirelessnumber.Numbercode

	return number
}
func (cgds *ContactGroupDataSource) FindLindlineNumber(numbercode string) string {

	lindlinenumber := LindlineNumber{}

	if err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C("landlinenumbers").Find(bson.M{"landlinenumber": numbercode}).One(&lindlinenumber); err != nil {
		return ""
	}

	number := lindlinenumber.Numbercode

	return number
}

func (cgds *ContactGroupDataSource) SaveContactGroup(contactGroup *model.ContactGroup) error {
	// if there is no contact group id assign one

	if contactGroup.ID == "" {
		contactGroup.ID = bson.NewObjectId()
	}

	// update total number
	contactGroup.TotalNumbers = int32(len(contactGroup.Numbers))

	existingContactListObject, err := cgds.ContactGroupWithId(contactGroup.ID)
	if err != nil {
		// contact group doesn't exist, create new
		contactGroup.CreateDate = time.Now().UTC()
		contactGroup.UpdateDate = time.Now().UTC()
		// Write the contact group to mongo
		if err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).Insert(contactGroup); err != nil {
			return err
		}
	} else {
		// contact group exists
		contactGroup.UpdateDate = time.Now().UTC()
		if err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).UpdateId(existingContactListObject.ID, contactGroup); err != nil {
			return err
		}
	}
	return nil
}

func (cgds *ContactGroupDataSource) DeleteContactGroup(contactGroupObject *model.ContactGroup) error {
	// delete contact group object from database
	if err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).RemoveId(contactGroupObject.ID); err != nil {
		// contact group delete fail
		return err
	}
	return nil
}

func (cgds *ContactGroupDataSource) DeleteContactGroupWithContactListId(contactListId bson.ObjectId) error {
	// delete contact group objects from database

	if _, err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).RemoveAll(bson.M{"contactListId": contactListId}); err != nil {
		// sound object delete fail
		return err
	}
	return nil
}

func (cgds *ContactGroupDataSource) HasNumber(cgId bson.ObjectId, number string, limit int, page int) (model.ContactGroups, int, error) {
	contactGroup := model.ContactGroups{}
	query1 := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C("contactgroups").Find(bson.M{"_id": cgId, "numbers.number": number})
	query2 := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C("contactgroups").Find(bson.M{"_id": cgId, "numbers.number": number})
	if page < 1 || limit < 1 {

		if err := query1.All(&contactGroup); err != nil {
			return nil, 0, err
		}
	} else {
		if err := query2.Sort(number).Skip((page - 1) * limit).Limit(limit).All(&contactGroup); err != nil {
			return nil, 0, err
		}
	}

	count, er := query1.Count()
	if er != nil {
		return nil, 0, er
	}

	return contactGroup, count, nil

}

func (cgds *ContactGroupDataSource) PullNumberFromContactGroup(cgId bson.ObjectId, number string) (int, error) {

	query := bson.M{"_id": cgId}
	update := bson.M{"$pull": bson.M{"numbers": bson.M{"number": number}}}
	updateInfo, err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).UpdateAll(query, update)

	return updateInfo.Updated, err
}

func (cgds *ContactGroupDataSource) UpdateNumberCount(cgId bson.ObjectId, totalNumbers int32) error {
	query := bson.M{"_id": cgId}
	update := bson.M{"$set": bson.M{"totalNumbers": totalNumbers}}
	return cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).Update(query, update)
}

func (cgds *ContactGroupDataSource) IncrementContactGroupNumberCount(cgId bson.ObjectId) error {
	query := bson.M{"_id": cgId}
	update := bson.M{"$inc": bson.M{"totalNumbers": 1}}
	return cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).Update(query, update)
}

func (cgds *ContactGroupDataSource) DecrementContactGroupNumberCount(cgId bson.ObjectId) error {
	query := bson.M{"_id": cgId}
	update := bson.M{"$inc": bson.M{"totalNumbers": -1}}
	return cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).Update(query, update)
}

func (cgds *ContactGroupDataSource) ContactGroupCount(cgId bson.ObjectId) (int32, error) {
	query := bson.M{"_id": cgId}
	contactgroup := model.ContactGroup{}
	err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).Find(query).One(&contactgroup)
	if err != nil {
		return 0, err
	}
	return (contactgroup.TotalNumbers), err
}

func (cgds *ContactGroupDataSource) PushInContactGroup(cgId bson.ObjectId, number string) error {
	var object TelnyxNumberObject
	if object.ID == "" {
		object.ID = bson.NewObjectId()
	}
	object.Number = number
	object.TimeZone = "PST"
	object.Numbertype = ""

	query := bson.M{"_id": cgId}
	update := bson.M{"$push": bson.M{"numbers": object}}
	return cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).Update(query, update)

}

func (cgds *ContactGroupDataSource) DeleteInContactGroup(object *model.TelnyxContactGroup) error {
	// delete contact group object from database
	query := bson.M{"_id": object.ContactGroupID}
	update := bson.M{"$pull": bson.M{"numbers": bson.M{"number": object.Number}}}
	if err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).Update(query, update); err != nil {
		// contact group delete fail
		return err
	}
	return nil
}

func (cgds *ContactGroupDataSource) CheckNumberInContactGroup(object bson.ObjectId, number string) (bool, error) {
	// delete contact group object from database
	var objectt model.ContactGroup
	query := bson.M{"_id": object, "numbers": bson.M{"$elemMatch": bson.M{"number": number}}}
	//update := bson.M{"$pull": bson.M{"numbers": bson.M{"number": object.Number} }}
	err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(cContactGroups).Find(query).One(&objectt)
	if err != nil {
		// contact group delete fail
		return false, err
	}

	return true, nil
}
