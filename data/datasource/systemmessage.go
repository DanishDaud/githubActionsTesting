package datasource

import (
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const csystemmessage = "systemmessage"

type SystemMessageDataSource struct {
	DataSource
}

// this method returns SystemMessageObject database object using object id
func (uds *SystemMessageDataSource) SystemMessageObject(messagetype int) (*model.SystemMessages, error) {
	systemmessage := model.SystemMessages{}

	err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(csystemmessage).Find(bson.M{"type": messagetype}).One(&systemmessage)

	if err != nil {
		return nil, err
	}
	return &systemmessage, err
}
