package service

import (
	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"gopkg.in/mgo.v2/bson"
)

type CustomFieldService struct {
	Service
}

func (smc *CustomFieldService) Save(obj model.CustomField) (*model.CustomField, error) {
	// get new instance of user data source
	sds := smc.ttsDataSource()
	defer sds.Session.Close()

	// get user object from database
	return sds.Save(obj)
}
func (smc *CustomFieldService) List(obj model.CustomFilter) (*[]model.CustomField, int, error) {
	// get new instance of user data source
	sds := smc.ttsDataSource()
	defer sds.Session.Close()

	// get user object from database
	return sds.List(obj)
}

func (smc *CustomFieldService) GetByID(objID string) (*model.CustomField, error) {
	// get new instance of user data source
	sds := smc.ttsDataSource()
	defer sds.Session.Close()

	objectid := bson.ObjectIdHex(objID)
	// get user object from database
	return sds.GetById(objectid)
}
func (smc *CustomFieldService) ttsDataSource() *datasource.CustomFieldDataSource {
	return &datasource.CustomFieldDataSource{DataSource: datasource.DataSource{Session: smc.Session.Copy()}}
}
