package datasource

import (
	"time"

	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const cBalance = "balance"

// Data Structure to represent Campaign
type BalanceDataSource struct {
	DataSource
}

func (bds *BalanceDataSource) Save(object *model.Balance) error {
	// if there is no campaign id assign one
	if object.ID == "" {
		object.ID = bson.NewObjectId()
		object.CreateDate = time.Now().UTC()
		object.UpdateDate = time.Now().UTC()
		// Write the campaign to mongo
		return bds.DbSession().DB(cmlutils.DefaultDatabase()).C(cBalance).Insert(object)
	} else {
		object.UpdateDate = time.Now().UTC()
		// Write the campaign to mongo
		return bds.DbSession().DB(cmlutils.DefaultDatabase()).C(cBalance).UpdateId(object.ID, object)
	}
}

func (bds *BalanceDataSource) GetGhostBalances(userid bson.ObjectId) (model.Balances, error) {
	balans := model.Balances{}
	err := bds.DbSession().DB(cmlutils.DefaultDatabase()).C(cBalance).Find(bson.M{"userid": userid, "isGhost": true}).All(&balans)
	return balans, err
}
