package service

import (
	"errors"

	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"gopkg.in/mgo.v2/bson"
)

type WepayCheckoutService struct {
	Service
}

// returns checkout database object of present in database
// otherwise it returns error
func (wcs *WepayCheckoutService) Object(objectID string) (*model.WepayCheckout, error) {
	// get new instance of user data source
	uds := wcs.wepaycheckoutDataSource()
	defer uds.Session.Close()

	if !bson.IsObjectIdHex(objectID) {
		return nil, errors.New("object id is not valid")
	}

	id := bson.ObjectIdHex(objectID)

	// get user object from database
	return uds.Object(id)
}

func (wcs *WepayCheckoutService) WepayCheckoutId(id int) (*model.WepayCheckout, error) {
	// get new instance of user data source
	uds := wcs.wepaycheckoutDataSource()
	defer uds.Session.Close()

	// get user object from database
	return uds.WepayCheckoutId(id)
}

func (wcs *WepayCheckoutService) Save(info *model.WepayCheckout) error {
	// get new instance of user data source
	uds := wcs.wepaycheckoutDataSource()
	defer uds.Session.Close()

	return uds.Save(info)
}

// return instance if user  data source
// every time a new instance would be created
func (wcs *WepayCheckoutService) wepaycheckoutDataSource() *datasource.WepayCheckoutDataSource {
	return &datasource.WepayCheckoutDataSource{DataSource: datasource.DataSource{Session: wcs.Session.Copy()}}
}
