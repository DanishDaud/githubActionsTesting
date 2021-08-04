package service

import (
	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"gopkg.in/mgo.v2/bson"
)

// Data Structure to represent balance service
type BalanceService struct {
	Service
}

func (bs *BalanceService) Save(object *model.Balance) error {
	// get new instance of sound file datasource
	bds := bs.balanceDatasource()
	defer bds.Session.Close()

	return bds.Save(object)
}

func (bs *BalanceService) GetNegBalance(userid bson.ObjectId) (model.Balances, float32, error) {
	// get new instance of sound file datasource
	bds := bs.balanceDatasource()
	defer bds.Session.Close()

	balns, err := bds.GetGhostBalances(userid)
	if err != nil {
		return nil, -1, err
	}

	var amount float32 = 0.0
	for _, bal := range balns {
		amount = amount + bal.Consumed
	}

	return nil, amount, nil
}

func (bs *BalanceService) balanceDatasource() *datasource.BalanceDataSource {
	return &datasource.BalanceDataSource{DataSource: datasource.DataSource{Session: bs.Session.Copy()}}
}
