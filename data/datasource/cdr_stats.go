package datasource

import (
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"time"

	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

func (cdrds *CdrDataSource) Aleg(userid bson.ObjectId, campId bson.ObjectId, cday time.Time) ([]model.Aleg, error) {
	var userDisposition = []model.Aleg{}
	var andQuery []map[string]interface{}
	if campId == "" {

		q0 := bson.M{
			"userid": userid,
		}
		andQuery = append(andQuery, q0)
		q2 := bson.M{
			"created_date": cday,
		}
		andQuery = append(andQuery, q2)

		err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(alegStats).Find(bson.M{"$and": andQuery}).All(&userDisposition)
		if err != nil {
			return nil, err
		}
		return userDisposition, nil
	} else {

		q0 := bson.M{
			"userid": userid,
		}
		andQuery = append(andQuery, q0)
		q1 := bson.M{
			"campaignid": campId,
		}

		andQuery = append(andQuery, q1)
		q2 := bson.M{
			"created_date": cday,
		}
		andQuery = append(andQuery, q2)

		err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(alegStats).Find(bson.M{"$and": andQuery}).All(&userDisposition)
		if err != nil {
			return nil, err
		}
		return userDisposition, nil
	}

}

func (cdrds *CdrDataSource) Bleg(userid bson.ObjectId, campId bson.ObjectId, cday time.Time) ([]model.Bleg, error) {
	var userDisposition = []model.Bleg{}

	var andQuery []map[string]interface{}
	if campId == "" {
		q0 := bson.M{
			"userid": userid,
		}
		andQuery = append(andQuery, q0)
		q2 := bson.M{
			"created_date": cday,
		}
		andQuery = append(andQuery, q2)

		err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(blegStats).Find(bson.M{"$and": andQuery}).All(&userDisposition)
		if err != nil {
			return nil, err
		}
		return userDisposition, nil
	} else {
		q0 := bson.M{
			"userid": userid,
		}
		andQuery = append(andQuery, q0)
		q1 := bson.M{
			"campaignid": campId,
		}

		andQuery = append(andQuery, q1)
		q2 := bson.M{
			"created_date": cday,
		}
		andQuery = append(andQuery, q2)

		err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(blegStats).Find(bson.M{"$and": andQuery}).All(&userDisposition)
		if err != nil {
			return nil, err
		}
		return userDisposition, nil
	}

}

func (cdrds *CdrDataSource) TotalDialed(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int

	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":     userid,
			"calltype":   ctype,
			"campaignId": campId,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":   userid,
			"calltype": ctype,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total_count"]
	}

	return count, err
}

func (cdrds *CdrDataSource) BillingInc(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int

	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":     userid,
			"calltype":   ctype,
			"campaignId": campId,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":   userid,
			"calltype": ctype,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$group": bson.M{
				"_id": "null",
				"total": bson.M{
					"$sum": "$billIncr",
				},
			},
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}
func (cdrds *CdrDataSource) UserAleg(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time) ([]model.Aleg, error) {
	var userDisposition = []model.Aleg{}
	var andQuery []map[string]interface{}
	var q2 bson.M
	if campId == "" {

		q0 := bson.M{
			"userid": userid,
		}
		andQuery = append(andQuery, q0)
		q2 = bson.M{
			"created_date": bson.M{
				"$gte": from,
				"$lte": to,
			},
		}
		andQuery = append(andQuery, q2)
		logrus.Infoln(andQuery)

		err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(alegStats).Find(bson.M{"$and": andQuery}).All(&userDisposition)
		if err != nil {
			return nil, err
		}
		return userDisposition, nil
	} else {

		q0 := bson.M{
			"userid": userid,
		}
		andQuery = append(andQuery, q0)
		q1 := bson.M{
			"campaignid": campId,
		}

		andQuery = append(andQuery, q1)

		q2 = bson.M{
			"created_date": bson.M{
				"$gte": from,
				"$lte": to,
			},
		}
		andQuery = append(andQuery, q2)
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(alegStats).Find(bson.M{"$and": andQuery}).All(&userDisposition)
	if err != nil {
		return nil, err
	}
	return userDisposition, nil
}

func (cdrds *CdrDataSource) UserBleg(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time) ([]model.Bleg, error) {
	var userDisposition = []model.Bleg{}
	var andQuery []map[string]interface{}
	var q2 bson.M
	if campId == "" {

		q0 := bson.M{
			"userid": userid,
		}
		andQuery = append(andQuery, q0)
		q2 = bson.M{
			"created_date": bson.M{
				"$gte": from,
				"$lte": to,
			},
		}
		andQuery = append(andQuery, q2)
		logrus.Infoln(andQuery)

		err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(blegStats).Find(bson.M{"$and": andQuery}).All(&userDisposition)
		if err != nil {
			return nil, err
		}
		return userDisposition, nil
	} else {

		q0 := bson.M{
			"userid": userid,
		}
		andQuery = append(andQuery, q0)
		q1 := bson.M{
			"campaignid": campId,
		}

		andQuery = append(andQuery, q1)

		q2 = bson.M{
			"created_date": bson.M{
				"$gte": from,
				"$lte": to,
			},
		}
		andQuery = append(andQuery, q2)
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(blegStats).Find(bson.M{"$and": andQuery}).All(&userDisposition)
	if err != nil {
		return nil, err
	}
	return userDisposition, nil
}

func (cdrds *CdrDataSource) AlegLastHour(userid bson.ObjectId, campId bson.ObjectId, from time.Time, fromhour int32, tohour int32) ([]model.AlegHourly, error) {
	var userDisposition = []model.AlegHourly{}
	var andQuery []map[string]interface{}
	var q2 bson.M
	var q3 bson.M
	logrus.Infoln(from)
	if campId == "" {

		q0 := bson.M{
			"userid": userid,
		}
		andQuery = append(andQuery, q0)
		q2 = bson.M{
			"created_date": from,
		}
		andQuery = append(andQuery, q2)
		q3 = bson.M{
			"from": bson.M{
				"$gte": fromhour,
				"$lte": tohour,
			},
		}
		andQuery = append(andQuery, q3)
		err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(alegStatshourly).Find(bson.M{"$and": andQuery}).All(&userDisposition)
		if err != nil {
			return nil, err
		}
		return userDisposition, nil
	} else {

		q0 := bson.M{
			"userid": userid,
		}
		andQuery = append(andQuery, q0)
		q1 := bson.M{
			"campaignid": campId,
		}

		andQuery = append(andQuery, q1)

		q2 = bson.M{
			"created_date": from,
		}
		andQuery = append(andQuery, q2)
		q3 = bson.M{
			"from": bson.M{
				"$gte": fromhour,
				"$lte": tohour,
			},
		}
		andQuery = append(andQuery, q3)
		logrus.Infoln(andQuery)
		err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(alegStatshourly).Find(bson.M{"$and": andQuery}).All(&userDisposition)
		if err != nil {
			return nil, err
		}
		return userDisposition, nil
	}

}

func (cdrds *CdrDataSource) AlegHourly(userid bson.ObjectId, campId bson.ObjectId, from time.Time, fromhour int32) ([]model.AlegHourly, error) {
	var userDisposition = []model.AlegHourly{}
	var andQuery []map[string]interface{}
	var q2 bson.M
	var q3 bson.M
	logrus.Infoln(from, fromhour)
	if campId == "" {

		q0 := bson.M{
			"userid": userid,
		}
		andQuery = append(andQuery, q0)
		q2 = bson.M{
			"created_date": from,
		}
		andQuery = append(andQuery, q2)
		q3 = bson.M{
			"from": fromhour,
		}
		andQuery = append(andQuery, q3)
		err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(alegStatshourly).Find(bson.M{"$and": andQuery}).All(&userDisposition)
		if err != nil {
			return nil, err
		}
		return userDisposition, nil
	} else {

		q0 := bson.M{
			"userid": userid,
		}
		andQuery = append(andQuery, q0)
		q1 := bson.M{
			"campaignid": campId,
		}

		andQuery = append(andQuery, q1)

		q2 = bson.M{
			"created_date": from,
		}
		andQuery = append(andQuery, q2)
		q3 = bson.M{
			"from": fromhour,
		}
		andQuery = append(andQuery, q3)
		logrus.Infoln(andQuery)
		err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(alegStatshourly).Find(bson.M{"$and": andQuery}).All(&userDisposition)
		if err != nil {
			return nil, err
		}
		return userDisposition, nil
	}

}

func (cdrds *CdrDataSource) UserDisposition(userid bson.ObjectId, from int, to int, cday time.Time) ([]model.CallInformation, error) {
	var userDisposition = []model.CallInformation{}
	var andQuery []map[string]interface{}
	q0 := bson.M{
		"userid": userid,
	}
	andQuery = append(andQuery, q0)
	q1 := bson.M{
		"to":   to,
		"from": from,
	}

	andQuery = append(andQuery, q1)
	q2 := bson.M{
		"created_date": cday,
	}
	andQuery = append(andQuery, q2)

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(uDispositionStats).Find(bson.M{"$and": andQuery}).All(&userDisposition)
	if err != nil {
		return nil, err
	}

	return userDisposition, nil
}

func (cdrds *CdrDataSource) AnswerCall(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int

	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":      userid,
			"calltype":    ctype,
			"disposition": 1,
			"campaignId":  campId,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":      userid,
			"calltype":    ctype,
			"disposition": 1,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) BusyCall(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int

	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":      userid,
			"calltype":    ctype,
			"disposition": 2,
			"campaignId":  campId,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":      userid,
			"calltype":    ctype,
			"disposition": 2,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) NoAnswerCall(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int

	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":      userid,
			"calltype":    ctype,
			"disposition": 3,
			"campaignId":  campId,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":      userid,
			"calltype":    ctype,
			"disposition": 3,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) Error(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int

	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":      userid,
			"calltype":    ctype,
			"disposition": 3,
			"campaignId":  campId,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":      userid,
			"calltype":    ctype,
			"disposition": 3,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) DNC(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int

	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":      userid,
			"calltype":    ctype,
			"disposition": 3,
			"campaignId":  campId,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":      userid,
			"calltype":    ctype,
			"disposition": 3,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) LiveCall(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int
	var andQuery []map[string]interface{}
	var match bson.M
	q0 := bson.M{
		"disposition": 1,
	}
	andQuery = append(andQuery, q0)
	q1 := bson.M{
		"amdStatus": 1,
	}
	andQuery = append(andQuery, q1)
	if campId != "" {
		match = bson.M{
			"userid":     userid,
			"calltype":   ctype,
			"$and":       andQuery,
			"campaignId": campId,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":   userid,
			"calltype": ctype,
			"$and":     andQuery,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) Failed(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int

	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":      userid,
			"calltype":    ctype,
			"disposition": 3,
			"campaignId":  campId,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":      userid,
			"calltype":    ctype,
			"disposition": 3,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) XAns(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int
	var andQuery []map[string]interface{}
	var match bson.M
	q0 := bson.M{
		"disposition": 1,
	}
	andQuery = append(andQuery, q0)
	q1 := bson.M{
		"calltype": 2,
	}
	andQuery = append(andQuery, q1)
	if campId != "" {
		match = bson.M{
			"userid":     userid,
			"calltype":   ctype,
			"$and":       andQuery,
			"campaignId": campId,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":   userid,
			"calltype": ctype,
			"$and":     andQuery,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) VM(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int
	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":     userid,
			"calltype":   ctype,
			"amdStatus":  2,
			"campaignId": campId,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":    userid,
			"calltype":  ctype,
			"amdStatus": 2,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) CallCostAll(payload model.GetCallCostPayload) (float32, error) {
	var result []map[string]float32
	var count float32

	pipiline := []bson.M{
		{
			"$match": bson.M{
				"userid": payload.UserId,
				"created_date": bson.M{
					"$gte": payload.From,
					"$lt":  payload.To,
				},
			},
		}, {
			"$group": bson.M{
				"_id": "null",
				"total": bson.M{
					"$sum": "$cost",
				},
			},
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(alegStats).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}
func (cdrds *CdrDataSource) CallMin(userid bson.ObjectId, from time.Time, to time.Time) (float32, error) {
	var result []map[string]float32
	var count float32

	pipiline := []bson.M{
		{
			"$match": bson.M{
				"userid": userid,
				"createDate": bson.M{
					"$gte": from,
					"$lt":  to,
				},
			},
		}, {
			"$group": bson.M{
				"_id": "null",
				"total": bson.M{
					"$sum": "$billDuration",
				},
			},
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) CallCount(userid bson.ObjectId, from time.Time, to time.Time) (float32, error) {
	var result []map[string]float32
	var count float32
	var match bson.M
	match = bson.M{
		"userid": userid,
		"createDate": bson.M{
			"$gte": from,
			"$lt":  to,
		},
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}
func (cdrds *CdrDataSource) CallStats(campaignid bson.ObjectId, contactid bson.ObjectId, ctype int) ([]model.FromNumber, error) {
	var andQuery []map[string]interface{}
	var result []model.FromNumber
	var finalresult []model.FromNumber
	q0 := bson.M{
		"disposition": ctype,
	}
	andQuery = append(andQuery, q0)
	q1 := bson.M{
		"campaignId": campaignid,
	}
	andQuery = append(andQuery, q1)
	page := 1
	limit := 10000
	totalFetched := 0
	if ctype == model.LiveStatus {
		query := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Find(bson.M{"campaignId": campaignid, "amdStatus": ctype})
		query1 := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Find(bson.M{"campaignId": campaignid, "amdStatus": ctype})
		count, er := query1.Count()
		if er != nil {
			return nil, er
		}
		for {
			if err := query.Sort("-$natural").Skip((page - 1) * limit).Limit(limit).All(&result); err != nil {
				return nil, err
			}
			finalresult = append(finalresult, result...)
			totalFetched = totalFetched + len(result)
			page = page + 1
			if totalFetched >= count {
				break
			}
		}

		return finalresult, nil
	} else if ctype == model.BusyStatus {
		query := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Find(bson.M{"campaignId": campaignid, "disposition": ctype})
		query1 := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Find(bson.M{"campaignId": campaignid, "disposition": ctype})
		count, er := query1.Count()
		if er != nil {
			return nil, er
		}
		for {
			if err := query.Sort("-$natural").Skip((page - 1) * limit).Limit(limit).All(&result); err != nil {
				return nil, err
			}
			finalresult = append(finalresult, result...)
			totalFetched = totalFetched + len(result)
			page = page + 1
			if totalFetched >= count {
				break
			}
		}

		return finalresult, nil
	} else if ctype == model.NoAnswer {
		query := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Find(bson.M{"campaignId": campaignid, "disposition": ctype})
		query1 := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Find(bson.M{"campaignId": campaignid, "disposition": ctype})
		count, er := query1.Count()
		if er != nil {
			return nil, er
		}
		for {
			if err := query.Sort("-$natural").Skip((page - 1) * limit).Limit(limit).All(&result); err != nil {
				return nil, err
			}
			finalresult = append(finalresult, result...)
			totalFetched = totalFetched + len(result)
			page = page + 1
			if totalFetched >= count {
				break
			}
		}

		return finalresult, nil
	} else if ctype == model.Failed {
		query := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Find(bson.M{"campaignId": campaignid, "disposition": ctype})
		query1 := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Find(bson.M{"campaignId": campaignid, "disposition": ctype})

		count, er := query1.Count()
		if er != nil {
			return nil, er
		}
		for {
			if err := query.Sort("-$natural").Skip((page - 1) * limit).Limit(limit).All(&result); err != nil {
				return nil, err
			}
			finalresult = append(finalresult, result...)
			totalFetched = totalFetched + len(result)
			page = page + 1
			if totalFetched >= count {
				break
			}
		}

		return finalresult, nil
	} else if ctype == model.Voicemail {
		query := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Find(bson.M{"campaignId": campaignid, "amdStatus": model.AMDStatusVM})
		query1 := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Find(bson.M{"campaignId": campaignid, "amdStatus": model.AMDStatusVM})

		count, er := query1.Count()
		if er != nil {
			return nil, er
		}
		for {
			if err := query.Sort("-$natural").Skip((page - 1) * limit).Limit(limit).All(&result); err != nil {
				return nil, err
			}
			finalresult = append(finalresult, result...)
			totalFetched = totalFetched + len(result)
			page = page + 1
			if totalFetched >= count {
				break
			}
		}

		return finalresult, nil

	}
	return nil, nil
}

func (cdrds *CdrDataSource) CallAlegCount(userid bson.ObjectId, from time.Time, to time.Time) (float32, error) {
	var result []map[string]float32
	var count float32
	var match bson.M
	match = bson.M{
		"userid": userid,
		"type":   1,
		"createDate": bson.M{
			"$gte": from,
			"$lt":  to,
		},
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) CallBlegCount(userid bson.ObjectId, from time.Time, to time.Time) (float32, error) {
	var result []map[string]float32
	var count float32
	var match bson.M
	match = bson.M{
		"userid": userid,
		"type":   2,
		"createDate": bson.M{
			"$gte": from,
			"$lt":  to,
		},
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) BillingDuration(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int

	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":     userid,
			"calltype":   ctype,
			"campaignId": campId,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":   userid,
			"calltype": ctype,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$group": bson.M{
				"_id": "null",
				"total": bson.M{
					"$sum": "$billDuration",
				},
			},
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) Cost(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int

	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":     userid,
			"calltype":   ctype,
			"campaignId": campId,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":   userid,
			"calltype": ctype,
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$group": bson.M{
				"_id": "null",
				"total": bson.M{
					"$sum": "$cost",
				},
			},
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) SixteyMin(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int

	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":     userid,
			"calltype":   ctype,
			"campaignId": campId,
			"billDuration": bson.M{
				"$gte": 30,
				"$lt":  60,
			},
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":   userid,
			"calltype": ctype,
			"billDuration": bson.M{
				"$gte": 30,
				"$lt":  60,
			},
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) ThirtyMin(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int

	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":     userid,
			"calltype":   ctype,
			"campaignId": campId,
			"billDuration": bson.M{
				"$gte": 20,
				"$lt":  30,
			},
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":   userid,
			"calltype": ctype,
			"billDuration": bson.M{
				"$gte": 20,
				"$lt":  30,
			},
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}
	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) TwentyMin(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int

	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":     userid,
			"calltype":   ctype,
			"campaignId": campId,
			"billDuration": bson.M{
				"$gte": 10,
				"$lt":  20,
			},
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":   userid,
			"calltype": ctype,
			"billDuration": bson.M{
				"$gte": 10,
				"$lt":  20,
			},
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) TenMin(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int

	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":     userid,
			"calltype":   ctype,
			"campaignId": campId,
			"billDuration": bson.M{
				"$gte": 5,
				"$lt":  10,
			},
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":   userid,
			"calltype": ctype,
			"billDuration": bson.M{
				"$gte": 5,
				"$lt":  10,
			},
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}

func (cdrds *CdrDataSource) FiveMin(userid bson.ObjectId, campId bson.ObjectId, from time.Time, to time.Time, ctype int) (int, error) {
	var result []map[string]int
	var count int

	var match bson.M

	if campId != "" {
		match = bson.M{
			"userid":     userid,
			"calltype":   ctype,
			"campaignId": campId,
			"billDuration": bson.M{
				"$gte": 0,
				"$lt":  5,
			},
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	} else {
		match = bson.M{
			"userid":   userid,
			"calltype": ctype,
			"billDuration": bson.M{
				"$gte": 0,
				"$lt":  5,
			},
			"createDate": bson.M{
				"$gte": from,
				"$lt":  to,
			},
		}
	}

	pipiline := []bson.M{
		{
			"$match": match,
		}, {
			"$count": "total_count",
		},
	}

	err := cdrds.DbSession().DB(cmlutils.DefaultDatabase()).C(cCdr).Pipe(pipiline).All(&result)
	if err != nil {
		return 0, err
	}

	for _, i := range result {
		count = i["total"]
	}

	return count, err
}
