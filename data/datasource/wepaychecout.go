package datasource

import (
	"time"

	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const cWepayCheckout = "wepaycheckouts"

// Data Structure to represent Campaign
type WepayCheckoutDataSource struct {
	DataSource
}

// this method returns user database object using object id
func (wcds *WepayCheckoutDataSource) Object(objectID bson.ObjectId) (*model.WepayCheckout, error) {
	info := model.WepayCheckout{}

	err := wcds.DbSession().DB(cmlutils.DefaultDatabase()).C(cWepayCheckout).FindId(objectID).One(&info)

	if err != nil {
		return nil, err
	}
	return &info, err
}

// this method returns user database object using object id
func (wcds *WepayCheckoutDataSource) WepayCheckoutId(id int) (*model.WepayCheckout, error) {
	info := model.WepayCheckout{}

	err := wcds.DbSession().DB(cmlutils.DefaultDatabase()).C(cWepayCheckout).Find(bson.M{"checkout_id": id}).One(&info)

	if err != nil {
		return nil, err
	}
	return &info, err
}

func (wcds *WepayCheckoutDataSource) Save(info *model.WepayCheckout) error {
	// if there is no user id assign onex
	if info.ID == "" {
		info.ID = bson.NewObjectId()
	}

	existingObject, err := wcds.Object(info.ID)
	if err != nil {
		// user doesn't exist, create new
		info.CreateDate = time.Now().UTC()
		info.UpdateDate = time.Now().UTC()
		// Write the user to mongo
		if err := wcds.DbSession().DB(cmlutils.DefaultDatabase()).C(cWepayCheckout).Insert(info); err != nil {
			return err
		}
	} else {
		// user exists
		info.UpdateDate = time.Now().UTC()
		if err := wcds.DbSession().DB(cmlutils.DefaultDatabase()).C(cWepayCheckout).UpdateId(existingObject.ID, info); err != nil {
			return err
		}
	}
	return nil
}
