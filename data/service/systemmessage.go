package service

import (
	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
)

type SystemMessageService struct {
	Service
}

func (smc *SystemMessageService) SystemMessageObject(messagetype int) (*model.SystemMessages, error) {
	// get new instance of user data source
	sds := smc.systemmessageDataSource()
	defer sds.Session.Close()

	// get user object from database
	return sds.SystemMessageObject(messagetype)
}

func (smc *SystemMessageService) systemmessageDataSource() *datasource.SystemMessageDataSource {
	return &datasource.SystemMessageDataSource{DataSource: datasource.DataSource{Session: smc.Session.Copy()}}
}
