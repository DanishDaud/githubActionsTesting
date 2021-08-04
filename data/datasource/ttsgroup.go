package datasource

import (
	"errors"
	"time"

	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const ttsContactGroups = "ttscontactgroups"

// Data Structure to represent ContactListDatasource
type TTSGroupDataSource struct {
	DataSource
}

func (cgds *TTSGroupDataSource) ContactGroupWithId(objectId bson.ObjectId) (*model.TTSListContactGroup, error) {
	contactGroup := model.TTSListContactGroup{}
	err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(ttsContactGroups).FindId(objectId).One(&contactGroup)
	if err != nil {
		return nil, errors.New("can not find contact group")
	}

	return &contactGroup, nil
}

func (cgds *TTSGroupDataSource) SaveContactGroup(contactGroup *model.TTSListContactGroup) error {
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
		if err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(ttsContactGroups).Insert(contactGroup); err != nil {
			return err
		}
	} else {
		// contact group exists
		contactGroup.UpdateDate = time.Now().UTC()
		if err := cgds.DbSession().DB(cmlutils.DefaultDatabase()).C(ttsContactGroups).UpdateId(existingContactListObject.ID, contactGroup); err != nil {
			return err
		}
	}
	return nil
}
