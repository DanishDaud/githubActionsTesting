package service

import (
	"errors"

	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"gopkg.in/mgo.v2/bson"
)

// Data Structure to represent contact list service
type ContactGroupService struct {
	Service
}

func (cgs *ContactGroupService) ContactGroupWithId(objectID string) (*model.ContactGroup, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, errors.New("object id is not valid")
	}

	contactGroupObjectId := bson.ObjectIdHex(objectID)

	// get new instance of sound file datasource
	contactGroupDataSource := cgs.contactGroupDatasource()
	defer contactGroupDataSource.Session.Close()

	return contactGroupDataSource.ContactGroupWithId(contactGroupObjectId)
}

func (cgs *ContactGroupService) SaveContactGroup(contactGroup *model.ContactGroup) error {
	// get new instance of sound file datasource
	contactGroupDataSource := cgs.contactGroupDatasource()
	defer contactGroupDataSource.Session.Close()

	return contactGroupDataSource.SaveContactGroup(contactGroup)
}
func (cgs *ContactGroupService) ExtractTimeNumberFromUSNumber(number string) string {
	contactGroupDataSource := cgs.contactGroupDatasource()
	defer contactGroupDataSource.Session.Close()

	return contactGroupDataSource.Findtimezone(number)
}
func (cgs *ContactGroupService) ExtractWirelessNumber(number string) string {
	contactGroupDataSource := cgs.contactGroupDatasource()
	defer contactGroupDataSource.Session.Close()

	return contactGroupDataSource.FindWirelessNumber(number)
}
func (cgs *ContactGroupService) ExtractLindlineNumber(number string) string {
	contactGroupDataSource := cgs.contactGroupDatasource()
	defer contactGroupDataSource.Session.Close()

	return contactGroupDataSource.FindLindlineNumber(number)
}

func (cgs *ContactGroupService) DeleteTelnyxContactGroup(number *model.TelnyxContactGroup) error {
	// get new instance of sound file datasource
	contactGroupDataSource := cgs.contactGroupDatasource()
	defer contactGroupDataSource.Session.Close()

	return contactGroupDataSource.DeleteInContactGroup(number)
}

func (cgs *ContactGroupService) DeleteContactGroup(contactGroup *model.ContactGroup) error {
	// get new instance of sound file datasource
	contactGroupDataSource := cgs.contactGroupDatasource()
	defer contactGroupDataSource.Session.Close()

	return contactGroupDataSource.DeleteContactGroup(contactGroup)
}

// return instance if contact group  datasource
// every time a new instance would be created
func (cgs *ContactGroupService) contactGroupDatasource() *datasource.ContactGroupDataSource {
	return &datasource.ContactGroupDataSource{DataSource: datasource.DataSource{Session: cgs.Session.Copy()}}
}

func (cgs *ContactGroupService) ContactGroupCount(objectID bson.ObjectId) (int32, error) {

	// get new instance of sound file datasource
	contactGroupDataSource := cgs.contactGroupDatasource()
	defer contactGroupDataSource.Session.Close()

	return contactGroupDataSource.ContactGroupCount(objectID)
}

func (cgs *ContactGroupService) PushInContactGroup(objectID bson.ObjectId, number string) error {

	// get new instance of sound file datasource
	contactGroupDataSource := cgs.contactGroupDatasource()
	defer contactGroupDataSource.Session.Close()

	return contactGroupDataSource.PushInContactGroup(objectID, number)
}

func (cgs *ContactGroupService) IncrementContactGroupNumberCount(objectID bson.ObjectId) error {

	// get new instance of sound file datasource
	contactGroupDataSource := cgs.contactGroupDatasource()
	defer contactGroupDataSource.Session.Close()

	return contactGroupDataSource.IncrementContactGroupNumberCount(objectID)
}

func (cgs *ContactGroupService) DecrementContactGroupNumberCount(objectID bson.ObjectId) error {

	// get new instance of sound file datasource
	contactGroupDataSource := cgs.contactGroupDatasource()
	defer contactGroupDataSource.Session.Close()

	return contactGroupDataSource.IncrementContactGroupNumberCount(objectID)
}

func (cgs *ContactGroupService) CheckNumberInContactGroup(objectID bson.ObjectId, number string) (bool, error) {

	// get new instance of sound file datasource
	contactGroupDataSource := cgs.contactGroupDatasource()
	defer contactGroupDataSource.Session.Close()

	return contactGroupDataSource.CheckNumberInContactGroup(objectID, number)
}
