package service

import (
	"errors"
	"github.com/gomarkho/sas-rvm-provapi/data/datasource"

	"github.com/gomarkho/sas-rvm-provapi/model"
	"gopkg.in/mgo.v2/bson"
)

// Data Structure to represent contact list service
type TTSGroupService struct {
	Service
}

func (cgs *TTSGroupService) ContactGroupWithId(objectID string) (*model.TTSListContactGroup, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, errors.New("object id is not valid")
	}

	contactGroupObjectId := bson.ObjectIdHex(objectID)

	// get new instance of sound file datasource
	contactGroupDataSource := cgs.ttsDataSource()
	defer contactGroupDataSource.Session.Close()

	return contactGroupDataSource.ContactGroupWithId(contactGroupObjectId)
}

func (cgs *TTSGroupService) SaveContactGroup(contactGroup *model.TTSListContactGroup) error {
	// get new instance of sound file datasource
	contactGroupDataSource := cgs.ttsGroupDatasource()
	defer contactGroupDataSource.Session.Close()

	return contactGroupDataSource.SaveContactGroup(contactGroup)
}

func (cgs *TTSGroupService) ttsDataSource() *datasource.TTSGroupDataSource {
	return &datasource.TTSGroupDataSource{DataSource: datasource.DataSource{Session: cgs.Session.Copy()}}
}
