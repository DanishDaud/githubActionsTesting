package datasource

import (
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

func (uds *UserDataSource) TotalCost(userid bson.ObjectId) (float32, error) {
	var result []map[string]float32
	var amount float32
	var andQuery []map[string]interface{}
	q0 := bson.M{
		"call.scheduleSettings.id": bson.M{"$ne": ""},
	}
	andQuery = append(andQuery, q0)
	pipiline := []bson.M{
		{
			"$match": bson.M{
				"userid": userid,
				"$and":   andQuery,
			},
		}, {
			"$group": bson.M{
				"_id": "null",
				"totalAmount": bson.M{
					"$sum": "$cost.campaignCost",
				},
			},
		},
	}

	err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		amount = i["totalAmount"]
	}

	return amount, err
}

func (uds *UserDataSource) ActiveCampaign(userid bson.ObjectId) (int, error) {
	var result []map[string]int
	var count int
	var andQuery []map[string]interface{}
	q0 := bson.M{
		"call.scheduleSettings.id": bson.M{"$ne": ""},
	}
	andQuery = append(andQuery, q0)
	q1 := bson.M{
		"status": bson.M{"$ne": 4},
	}
	andQuery = append(andQuery, q1)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"userid": userid,
				"$and":   andQuery,
			},
		},
		{
			"$count": "total_count"},
	}

	err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCampaigns).Pipe(pipeline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total_count"]

	}

	return count, err
}
