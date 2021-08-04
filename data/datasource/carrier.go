package datasource

import (
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const Carrier = "carriers"

type CarrierDataSource struct {
	DataSource
}

// this method returns user database object using object id
func (uds *CarrierDataSource) Carrier(objectId bson.ObjectId) (*model.Carrier, error) {
	carrier := model.Carrier{}
	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(Carrier).
		Find(bson.M{"_id": objectId}).One(&carrier); err != nil {
		return nil, err
	}

	return &carrier, nil
}

// this method returns user database object using object id
func (uds *CarrierDataSource) DefaultCarrier() (*model.Carrier, error) {
	carrier := model.Carrier{}
	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(Carrier).
		Find(bson.M{"default": true}).One(&carrier); err != nil {
		return nil, err
	}

	return &carrier, nil
}
