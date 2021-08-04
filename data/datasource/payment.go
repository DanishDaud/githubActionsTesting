package datasource

import (
	"time"

	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const cPayment = "payments"

// Data Structure to represent Campaign
type PaymentDataSource struct {
	DataSource
}

func (pds *PaymentDataSource) Save(object *model.Payment) error {
	// if there is no campaign id assign one
	if object.ID == "" {
		object.ID = bson.NewObjectId()
	}

	// campaign doesn't exist, create new
	object.CreateDate = time.Now().UTC()
	object.UpdateDate = time.Now().UTC()
	// Write the campaign to mongo
	if err := pds.DbSession().DB(cmlutils.DefaultDatabase()).C(cPayment).Insert(object); err != nil {
		return err
	}

	return nil
}
func (pds *PaymentDataSource) GetPaymentObject(id bson.ObjectId, page int, limit int) (model.Payments, int, error) {
	// if there is no campaign id assign one
	payemt := model.Payments{}

	query := pds.DbSession().DB(cmlutils.DefaultDatabase()).C(cPayment).Find(bson.M{"userid": id})
	query1 := pds.DbSession().DB(cmlutils.DefaultDatabase()).C(cPayment).Find(bson.M{"userid": id})

	if page < 1 || limit < 1 {
		if err := query.All(&payemt); err != nil {
			return nil, 0, err
		}
	} else {
		if err := query.Sort("-createDate").Skip((page - 1) * limit).Limit(limit).All(&payemt); err != nil {
			return nil, 0, err
		}
	}

	count, er := query1.Count()
	if er != nil {
		return nil, 0, er
	}

	return payemt, count, nil

}
