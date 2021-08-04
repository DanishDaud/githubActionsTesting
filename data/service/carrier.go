package service

import (
	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"gopkg.in/mgo.v2/bson"
)

type CarrierService struct {
	Service
}

// returns user database object if present in database
// otherwise it returns error
func (us *CarrierService) CarrierObject(objectId bson.ObjectId) (*model.Carrier, error) {
	// get new instance of user data source
	uds := us.CarrierDataSource()
	defer uds.Session.Close()

	// get user object from database
	return uds.Carrier(objectId)
}

func (us *CarrierService) DefaultCarrierObject() (*model.Carrier, error) {
	// get new instance of user data source
	uds := us.CarrierDataSource()
	defer uds.Session.Close()

	// get user object from database
	return uds.DefaultCarrier()
}

// return instance if user  data source
// every time a new instance would be created
func (us *CarrierService) CarrierDataSource() *datasource.CarrierDataSource {
	return &datasource.CarrierDataSource{DataSource: datasource.DataSource{Session: us.Session.Copy()}}
}
