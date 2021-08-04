package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"

	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"gopkg.in/mgo.v2/bson"
)

// Data Structure to represent campaign service
type CdrService struct {
	Service
}

func (cdrs *CdrService) GetById(objectID string) (*model.CDR, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, errors.New("object id is not valid")
	}

	objId := bson.ObjectIdHex(objectID)

	cdrds := cdrs.cdrDataSource()
	defer cdrds.Session.Close()

	return cdrds.GetById(objId)
}

func (cdrs *CdrService) Save(obj *model.CDR) error {
	cdrds := cdrs.cdrDataSource()
	defer cdrds.Session.Close()

	return cdrds.Save(obj)
}

func (cdrs *CdrService) CdrList(objectID string, from string, to string, amdStatus int8, callStatus int8, callType int, campaign string, page int, limit int) (model.CdrList, int, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, 0, errors.New("User object id is not valid")
	}

	userObjectId := bson.ObjectIdHex(objectID)

	cdrds := cdrs.cdrDataSource()
	defer cdrds.Session.Close()

	return cdrds.CdrList(userObjectId, page, limit, from, to, int8(amdStatus), int8(callStatus), callType, campaign)
}

func (cdrs *CdrService) CdrListForCampaigns(cids []string, page int, limit int) (*model.CampaingList, int, error) {

	ids := []bson.ObjectId{}

	for _, i := range cids {
		if !bson.IsObjectIdHex(i) {
			return nil, 0, errors.New("Campaign object id is not valid")
		}

		ids = append(ids, bson.ObjectIdHex(i))
	}

	cdrds := cdrs.cdrDataSource()
	defer cdrds.Session.Close()

	return cdrds.CdrListForCampaigns(ids, page, limit)
}

func (cdrs *CdrService) CallStatsALeg(payload model.UserStasPayload) (*model.Aleg, error) {

	var xfer float32
	alegstats := model.Aleg{}
	cdrds := cdrs.cdrDataSource()
	defer cdrds.Session.Close()

	aleg, err := cdrds.UserAleg(payload.UserId, payload.CampaignId, payload.From, payload.To)
	if err != nil {
		return nil, err
	}

	for _, i := range aleg {
		alegstats.NoAnswer += i.NoAnswer
		alegstats.Answer += i.Answer
		alegstats.Live += i.Live
		alegstats.VM += i.VM
		alegstats.Cost += i.Cost
		alegstats.Dail += i.Dail
		xfer += float32(i.Xfer)

		alegstats.Xfer = 0
		if alegstats.Dail > 0 {
			alegstats.Xfer = (float32(xfer) / float32(alegstats.Dail)) * 100
		}

		alegstats.Liveper = 0
		if alegstats.Answer > 0 {
			alegstats.Liveper = (float32(alegstats.Live) / float32(alegstats.Answer)) * 100
		}

		alegstats.Vmper = 0.0
		if alegstats.Answer > 0 {
			alegstats.Vmper = (float32(alegstats.VM) / float32(alegstats.Answer)) * 100
		}

		alegstats.BillDuration += i.BillDuration / 60
		alegstats.IncrDur += i.IncrDur / 60

		alegstats.Aloc = 0
		if alegstats.Answer > 0 {
			alegstats.Aloc = float32(alegstats.BillDuration) / float32(alegstats.Answer)
		}

		alegstats.ASR = 0
		if alegstats.Dail > 0 {
			alegstats.ASR = (float32(alegstats.Answer) / float32(alegstats.Dail)) * 100
		}

		alegstats.CreatedDate = i.CreatedDate
		alegstats.Redail = 0
	}

	return &alegstats, nil

}

func (cdrs *CdrService) CallStatsBLeg(payload model.UserStasPayload) (*model.Bleg, *model.ListeningStats, error) {
	blegstats := model.Bleg{}

	listenstats := model.ListeningStats{}
	cdrds := cdrs.cdrDataSource()
	defer cdrds.Session.Close()

	bleg, err := cdrds.UserBleg(payload.UserId, payload.CampaignId, payload.From, payload.To)
	if err != nil {
		return nil, nil, err
	}
	for _, i := range bleg {
		blegstats.NoAnswer += i.NoAnswer
		blegstats.Answer += i.Answer
		blegstats.Cost += i.Cost
		blegstats.Dail += i.Dail
		blegstats.CreatedDate = i.CreatedDate
		blegstats.BillDuration += i.BillDuration / 60
		blegstats.IncrDur += i.IncrDur / 60
		blegstats.Aloc = float32(blegstats.BillDuration) / float32(blegstats.Answer)
		blegstats.ASR = (float32(blegstats.Answer) / float32(blegstats.Dail)) * 100
		listenstats.TwentyMin += int64(i.TwentyMin)
		listenstats.TenMin += int64(i.TenMin)
		listenstats.FiveMin += int64(i.FiveMin)
		listenstats.Sixty += int64(i.Sixty)
		listenstats.Thirty += int64(i.Thirty)
		listenstats.MoreTwentry += int64(i.MoreTwenty)

	}

	return &blegstats, &listenstats, nil

}

func (cdrs *CdrService) CallStats(campId string, contactid bson.ObjectId, status int) ([]model.FromNumber, error) {

	cdrds := cdrs.cdrDataSource()
	defer cdrds.Session.Close()
	var campIdObj bson.ObjectId

	if campId != "" && bson.IsObjectIdHex(campId) {
		campIdObj = bson.ObjectIdHex(campId)
	}
	number, err := cdrds.CallStats(campIdObj, contactid, status)
	if err != nil {
		return nil, err
	}
	return number, nil
}

func (cdrs *CdrService) BLegStats(userId string, campId string, ptype int) (*model.Bleg, *model.ListeningStats, error) {
	if !bson.IsObjectIdHex(userId) {
		//return nil, 0, errors.New("User object id is not valid")
	}
	defaultFormat := "2006-01-02"
	//  var aloc float32
	//var liveper float32
	//	var xfer float32
	//	var vmper float32
	//	var asr float32
	var campIdObj bson.ObjectId
	if campId != "" && bson.IsObjectIdHex(campId) {
		campIdObj = bson.ObjectIdHex(campId)
	}

	userObjectId := bson.ObjectIdHex(userId)
	blegstats := model.Bleg{}

	listenstats := model.ListeningStats{}
	cdrds := cdrs.cdrDataSource()
	defer cdrds.Session.Close()
	loc, _ := time.LoadLocation("America/Los_Angeles")
	now := time.Now().In(loc)
	//s:=time.Now().Format(defaultFormat)
	s := now.Format(defaultFormat)
	c_day, err1 := time.ParseInLocation(defaultFormat, s, loc)
	if err1 != nil {
		logrus.Errorln(err1.Error())
	}
	// day:=time.Now().Format(defaultFormat)
	//c_day,err:=time.Parse(defaultFormat,day)
	//if err!=nil{
	//logrus.Infoln(err.Error())
	// }
	//day:=time.Now().Format(defaultFormat)
	//c_day,err:=time.Parse(defaultFormat,day)
	//if err!=nil{
	//	logrus.Infoln(err.Error())
	//	}
	bleg, err := cdrds.Bleg(userObjectId, campIdObj, c_day)
	if err != nil {
		return nil, nil, err
	}

	for _, i := range bleg {
		blegstats.NoAnswer += i.NoAnswer
		blegstats.Answer += i.Answer
		blegstats.Cost += i.Cost
		blegstats.Dail += i.Dail
		blegstats.CreatedDate = i.CreatedDate
		blegstats.BillDuration += i.BillDuration / 60
		blegstats.IncrDur += i.IncrDur / 60
		blegstats.Aloc = float32(blegstats.BillDuration) / float32(blegstats.Answer)
		blegstats.ASR = (float32(blegstats.Answer) / float32(blegstats.Dail)) * 100
		listenstats.TwentyMin += int64(i.TwentyMin)
		listenstats.TenMin += int64(i.TenMin)
		listenstats.FiveMin += int64(i.FiveMin)
		listenstats.Sixty += int64(i.Sixty)
		listenstats.Thirty += int64(i.Thirty)
		listenstats.MoreTwentry += int64(i.MoreTwenty)

	}

	return &blegstats, &listenstats, nil
}

func (cdrs *CdrService) CallCost(payload model.GetCallCostPayload) (float32, error) {
	cdrds := cdrs.cdrDataSource()
	defer cdrds.Session.Close()

	if !bson.IsObjectIdHex(payload.UserId.Hex()) {
		return -1, errors.New("object id is not valid")
	}
	// get  object from database
	return cdrds.CallCostAll(payload)

}

func (cdrs *CdrService) AlegStats(userId string) ([]model.ALegCall, error) {
	var costs []model.ALegCall

	var useralegcount model.ALegCall
	if !bson.IsObjectIdHex(userId) {
		return nil, errors.New("invalid user id")
	}
	defaultFormat := "2006-01-02"
	userObjectId := bson.ObjectIdHex(userId)

	times := generate24HoursTime()

	cdrds := cdrs.cdrDataSource()
	defer cdrds.Session.Close()

	loc, _ := time.LoadLocation("America/Los_Angeles")
	for index, i := range times {
		from := i
		to := times[index]
		from_hour := from.Hour()
		to_hour := to.Hour() + 1
		day := i.Format(defaultFormat)
		cday, _ := time.ParseInLocation(defaultFormat, day, loc)
		Time := fmt.Sprintf("%d:00", i.Hour())
		list, err := cdrds.UserDisposition(userObjectId, from_hour, to_hour, cday)
		if err != nil {
			return nil, err
		}

		if len(list) == 0 {
			useralegcount.Time = Time
			useralegcount.Count = 0
			costs = append(costs, useralegcount)
		} else {
			for _, i := range list {
				useralegcount.Time = Time
				useralegcount.Count += i.AlegCount
				costs = append(costs, useralegcount)

			}

		}

	}

	return costs, nil
}

func (cdrs *CdrService) BlegStats(userId string) ([]model.ALegCall, error) {
	var costs []model.ALegCall

	var useralegcount model.ALegCall
	if !bson.IsObjectIdHex(userId) {
		return nil, errors.New("invalid user id")
	}
	defaultFormat := "2006-01-02"
	userObjectId := bson.ObjectIdHex(userId)

	times := generate24HoursTime()

	cdrds := cdrs.cdrDataSource()
	defer cdrds.Session.Close()

	loc, _ := time.LoadLocation("America/Los_Angeles")
	for index, i := range times {
		from := i

		to := times[index]
		from_hour := from.Hour()
		to_hour := to.Hour() + 1
		day := i.Format(defaultFormat)
		cday, _ := time.ParseInLocation(defaultFormat, day, loc)
		Time := fmt.Sprintf("%d:00", i.Hour())
		list, err := cdrds.UserDisposition(userObjectId, from_hour, to_hour, cday)
		if err != nil {
			return nil, err
		}

		if len(list) == 0 {
			useralegcount.Time = Time
			useralegcount.Count = 0
			costs = append(costs, useralegcount)
		} else {
			for _, i := range list {
				useralegcount.Time = Time
				useralegcount.Count += i.BlegCount
				costs = append(costs, useralegcount)

			}

		}

	}

	return costs, nil
}

func (cdrs *CdrService) UserCostStats(userId string) ([]model.CallCost, error) {
	var costs []model.CallCost
	var usercost model.CallCost
	if !bson.IsObjectIdHex(userId) {
		return nil, errors.New("invalid user id")
	}
	defaultFormat := "2006-01-02"
	userObjectId := bson.ObjectIdHex(userId)

	times := generate24HoursTime()

	cdrds := cdrs.cdrDataSource()
	defer cdrds.Session.Close()
	loc, _ := time.LoadLocation("America/Los_Angeles")

	for index, i := range times {
		from := i
		to := times[index]
		from_hour := from.Hour()
		to_hour := to.Hour() + 1
		day := i.Format(defaultFormat)
		cday, _ := time.ParseInLocation(defaultFormat, day, loc)
		Time := fmt.Sprintf("%d:00", i.Hour())
		list, err := cdrds.UserDisposition(userObjectId, from_hour, to_hour, cday)
		if err != nil {
			return nil, err
		}

		if len(list) == 0 {
			usercost.Time = Time
			usercost.Cost = 0
			costs = append(costs, usercost)
		} else {
			for _, i := range list {
				usercost.Time = Time
				usercost.Cost += i.Cost
				costs = append(costs, usercost)
			}

		}

	}

	return costs, nil
}
func (cdrs *CdrService) UserDispositionStats(userId string) ([]model.CallDispositions, error) {
	var costs []model.CallDispositions
	defaultFormat := "2006-01-02"
	var userdis model.CallDispositions
	if !bson.IsObjectIdHex(userId) {
		return nil, errors.New("invalid user id")
	}

	userObjectId := bson.ObjectIdHex(userId)

	times := generate24HoursTime()

	cdrds := cdrs.cdrDataSource()
	defer cdrds.Session.Close()
	loc, _ := time.LoadLocation("America/Los_Angeles")
	for index, i := range times {
		from := i
		to := times[index]
		from_hour := from.Hour()
		to_hour := to.Hour() + 1
		day := i.Format(defaultFormat)
		cday, _ := time.ParseInLocation(defaultFormat, day, loc)
		Time := fmt.Sprintf("%d:00", i.Hour())
		list, err := cdrds.UserDisposition(userObjectId, from_hour, to_hour, cday)
		if err != nil {
			return nil, err
		}

		if len(list) == 0 {
			userdis.Time = Time
			userdis.Answered = 0
			userdis.NoAnswer = 0
			userdis.Failed = 0
			userdis.DNC = 0
			userdis.Busy = 0
			userdis.VM = 0
			costs = append(costs, userdis)
		} else {
			for _, i := range list {
				userdis.Time = Time
				userdis.Answered = i.Answerd
				userdis.NoAnswer = i.NoAnswer
				userdis.Failed = i.Failed
				userdis.DNC = i.DNC
				userdis.Busy = i.Busy
				userdis.VM = i.VM

				costs = append(costs, userdis)
			}

		}

	}

	return costs, nil
}
func (cdrs *CdrService) UserCallMinStats(userId string) ([]model.CallCount, error) {
	var stats []model.CallCount
	var userstats model.CallCount
	if !bson.IsObjectIdHex(userId) {
		return nil, errors.New("invalid user id")
	}
	defaultFormat := "2006-01-02"
	userObjectId := bson.ObjectIdHex(userId)

	times := generate24HoursTime()

	cdrds := cdrs.cdrDataSource()
	defer cdrds.Session.Close()
	loc, _ := time.LoadLocation("America/Los_Angeles")
	for index, i := range times {
		from := i

		to := times[index]
		from_hour := from.Hour()
		to_hour := to.Hour() + 1
		day := i.Format(defaultFormat)
		cday, _ := time.ParseInLocation(defaultFormat, day, loc)

		Time := fmt.Sprintf("%d:00", i.Hour())
		list, err := cdrds.UserDisposition(userObjectId, from_hour, to_hour, cday)
		if err != nil {
			return nil, err
		}

		if len(list) == 0 {
			userstats.Time = Time
			userstats.Count = 0
			userstats.Min = 0
			stats = append(stats, userstats)
		} else {
			for _, i := range list {
				userstats.Time = Time
				userstats.Count = i.Answerd + i.NoAnswer + i.Failed + i.DNC + i.Busy + i.Timeout
				userstats.Min = i.BillDuration / 60
				stats = append(stats, userstats)
			}

		}

	}
	return stats, nil

}

func (cdrs *CdrService) UserCallLoc(userId string) ([]model.CallLoc, error) {
	var stats []model.CallLoc
	var userloc model.CallLoc
	if !bson.IsObjectIdHex(userId) {
		return nil, errors.New("invalid user id")
	}
	defaultFormat := "2006-01-02"
	userObjectId := bson.ObjectIdHex(userId)

	times := generate24HoursTime()

	cdrds := cdrs.cdrDataSource()
	defer cdrds.Session.Close()
	loc, _ := time.LoadLocation("America/Los_Angeles")
	for index, i := range times {
		from := i

		to := times[index]
		from_hour := from.Hour()
		to_hour := to.Hour() + 1
		day := i.Format(defaultFormat)
		cday, _ := time.ParseInLocation(defaultFormat, day, loc)

		Time := fmt.Sprintf("%d:00", i.Hour())
		list, err := cdrds.UserDisposition(userObjectId, from_hour, to_hour, cday)
		if err != nil {
			return nil, err
		}

		if len(list) == 0 {
			userloc.Time = Time
			userloc.ALOC = 0
			userloc.ASR = 0

			stats = append(stats, userloc)
		} else {
			for _, i := range list {
				userloc.Time = Time
				answer := i.Answerd
				billdur := i.BillDuration
				Totalcall := i.Answerd + i.Busy + i.DNC + i.Failed + i.NoAnswer + i.Timeout
				asr := (float64(answer) / float64(Totalcall)) * 100
				aloc := float64(billdur) / float64(answer)

				userloc.ASR = asr
				userloc.ALOC = aloc

				stats = append(stats, userloc)
			}

		}

	}

	return stats, nil
}

func generate24HoursTime() []time.Time {
	var times []time.Time
	defaultFormat := "2006-01-02"
	for i := 0; i <= 23; i++ {
		h := strconv.Itoa(i)
		loc, _ := time.LoadLocation("America/Los_Angeles")
		now := time.Now().In(loc)
		//s:=time.Now().Format(defaultFormat)
		s := now.Format(defaultFormat) + " " + h
		t, err1 := time.ParseInLocation("2006-01-02 15", s, loc)
		if err1 != nil {
			logrus.Errorln(err1.Error())
		}
		//s:=time.Now().Format(defaultFormat)+" "+h
		//	t,err:=time.Parse("2006-01-02 15",s)
		//		logrus.Infoln(t)
		//	if err!=nil{
		//		logrus.Infoln(err.Error())
		//	}

		times = append(times, t)
	}

	return times
}

func getFromToPtype(ptype int) (time.Time, time.Time) {
	to := time.Now()
	from := time.Now()

	// 1 = last 30 days
	// 2 = last 7 days
	// 3 = last 24 hours
	// 4 = last 3 hours
	// 5 = last 1 hour

	switch ptype {
	case 1: // last 30 days
		from = time.Now().Add(-time.Hour * 24 * 30)
	case 2: // last 7 days
		from = time.Now().Add(-time.Hour * 24 * 7)
	case 3: // last 24 hours
		from = time.Now().Add(-time.Hour * 24)
	case 4: // last 3 hours
		from = time.Now().Add(-time.Hour * 3)
	case 5: // last 1 hour
		from = time.Now().Add(-time.Hour * 1)
	default:
	}

	return from, to
}

func (cdrs *CdrService) cdrDataSource() *datasource.CdrDataSource {
	return &datasource.CdrDataSource{DataSource: datasource.DataSource{Session: cdrs.Session.Copy()}}
}
