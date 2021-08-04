package service

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"gopkg.in/mgo.v2/bson"
)

// Data Structure to represent balance service
type PaymentService struct {
	Service
}

func (ps *PaymentService) Save(object *model.Payment) error {
	// get new instance of sound file datasource
	pds := ps.paymentDatasource()
	defer pds.Session.Close()

	return pds.Save(object)
}
func (ps *PaymentService) GetPaymentObject(Id string, page int, limit int) (model.Payments, int, error) {
	// get new instance of sound file datasource
	pds := ps.paymentDatasource()
	defer pds.Session.Close()

	if !bson.IsObjectIdHex(Id) {
		return nil, 0, errors.New("object id is not valid")
	}

	objId := bson.ObjectIdHex(Id)
	logrus.Infoln(objId)
	return pds.GetPaymentObject(objId, page, limit)
}

func (bs *PaymentService) paymentDatasource() *datasource.PaymentDataSource {
	return &datasource.PaymentDataSource{DataSource: datasource.DataSource{Session: bs.Session.Copy()}}
}
