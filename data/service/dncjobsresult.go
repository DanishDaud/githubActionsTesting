package service

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

var dncJobResultCronProcessingIndicator = make(chan int, 1)

func init() {
	// initial refill
	dncJobResultCronProcessingIndicator <- 1
}

type DNCJobResultService struct {
	Service
	RedisService
}

func (djrs *DNCJobResultService) dncJobsDataSource() *datasource.DNCJobsDataSource {
	return &datasource.DNCJobsDataSource{DataSource: datasource.DataSource{Session: djrs.Session.Copy()}}
}

func (djrs *DNCJobResultService) dncJobResultDataSource() *datasource.DNCJobsResultDataSource {
	return &datasource.DNCJobsResultDataSource{DataSource: datasource.DataSource{Session: djrs.Session.Copy()}}
}

func (djrs *DNCJobResultService) contactListDataSource() *datasource.ContactListDataSource {
	return &datasource.ContactListDataSource{DataSource: datasource.DataSource{Session: djrs.Session.Copy()}}
}

func (djrs *DNCJobResultService) ttsListDataSource() *datasource.TTSListDataSource {
	return &datasource.TTSListDataSource{DataSource: datasource.DataSource{Session: djrs.Session.Copy()}}
}

func (djrs *DNCJobResultService) contactGroupDataSource() *datasource.ContactGroupDataSource {
	return &datasource.ContactGroupDataSource{DataSource: datasource.DataSource{Session: djrs.Session.Copy()}}
}

func (djrs *DNCJobResultService) ttsmetaDataSource() *datasource.TTSMetaDataSource {
	return &datasource.TTSMetaDataSource{DataSource: datasource.DataSource{Session: djrs.Session.Copy()}}
}

// return instance if contact group  datasource
// every time a new instance would be created
func (djrs *DNCJobResultService) ttsGroupDatasource() *datasource.TTSGroupDataSource {
	return &datasource.TTSGroupDataSource{DataSource: datasource.DataSource{Session: djrs.Session.Copy()}}
}

func (djrs *DNCJobResultService) dncJobResultObject(dncjobs *model.DNCJobResult) error {
	// get new instance of sound file datasource
	djrds := djrs.dncJobResultDataSource()
	defer djrds.Session.Close()

	return djrds.SaveDncJobResult(dncjobs)
}

func (djrs *DNCJobResultService) DeleteDncJobObject(dncjobs *model.DNCJobResult) error {
	// get new instance of contact list datasource
	djrds := djrs.dncJobResultDataSource()
	defer djrds.Session.Close()

	return djrds.DeleteDncJobResult(dncjobs)
}

func (djrs *DNCJobResultService) DncjobResultObjectWithId(objectID string) (*model.DNCJobResult, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, errors.New("object id is not valid")
	}

	dncjobObjectId := bson.ObjectIdHex(objectID)

	// get new instance of contact list datasource
	djrds := djrs.dncJobResultDataSource()
	defer djrds.Session.Close()
	return djrds.DncJobResultObjectWithId(dncjobObjectId)
}

func (djrs *DNCJobResultService) DncjobResultObjectWithJobId(status string) (*model.DNCJobResult, error) {

	// get new instance of contact list datasource
	djrds := djrs.dncJobResultDataSource()
	defer djrds.Session.Close()
	return djrds.DncJobResultObjectWithJobId(status)
}

func (djrs *DNCJobResultService) DncjobResultObjectWithDncJobId(status string) (*model.DNCJobResult, error) {

	// get new instance of contact list datasource
	djrds := djrs.dncJobResultDataSource()
	defer djrds.Session.Close()
	return djrds.DncJobResultObjectWithDncJobId(status)
}

func (djrs *DNCJobResultService) DncjobResultObjectWithContactListId(objectID string, page int, limit int) (*model.DNCJobsResultList, int, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, 0, errors.New("object id is not valid")
	}

	contactlistId := bson.ObjectIdHex(objectID)

	// get new instance of contact list datasource
	djrds := djrs.dncJobResultDataSource()
	defer djrds.Session.Close()
	return djrds.DncJobResultWithContactListId(contactlistId, page, limit)
}

func (djrs *DNCJobResultService) JobResultCountWithContactListId(objectID string) (int, error) {
	if !bson.IsObjectIdHex(objectID) {
		return 0, errors.New("object id is not valid")
	}

	contactlistId := bson.ObjectIdHex(objectID)
	djrds := djrs.dncJobResultDataSource()
	defer djrds.Session.Close()
	return djrds.DncJobResultCountWithContactListId(contactlistId)
}

func (djrs *DNCJobResultService) DNCJobResultCron() {
	select {
	case temp := <-dncJobResultCronProcessingIndicator:
		logrus.Debugln("DNCJobResultCron Started ", temp)
		djrs.ProcessDNCJobResultCron()
		dncJobResultCronProcessingIndicator <- 1
		logrus.Debugln("DNCJobResultCron Ended ", temp)
	default:
		logrus.Debugln("Processing channel occupied")
	}
}

func (djrs *DNCJobResultService) ProcessDNCJobResultCron() {
	// get new instance of dnc job datasource
	dncJobds := djrs.dncJobsDataSource()
	defer dncJobds.Session.Close()

	// get new instance of dnc job result datasource
	dncJobResds := djrs.dncJobResultDataSource()
	defer dncJobResds.Session.Close()

	ttsListds := djrs.ttsListDataSource()
	defer ttsListds.Session.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		djrs.ProcessJobResultsForContactList()
	}()

	go func() {
		defer wg.Done()
		djrs.ProcessJobResultsForTTSList()
	}()

	wg.Wait()
}

func (djrs *DNCJobResultService) ProcessJobResultsForContactList() {
	dncJobds := djrs.dncJobsDataSource()
	defer dncJobds.Session.Close()

	contactListds := djrs.contactListDataSource()
	defer contactListds.Session.Close()

	dncJobResds := djrs.dncJobResultDataSource()
	defer dncJobResds.Session.Close()

	lists, _, err := contactListds.ContactListGetListWithDNCScrubStatus(model.DNCScrubProcessStatusAllBatchesSent, 1, 50)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("TTSListGetListWithDNCScrubStatus => Error :%s ::", err.Error()))
		return
	}

	for _, list := range *lists {
		jobsCount, err := dncJobds.DNCJobCountWithContactListId(list.ID)
		if err != nil {
			continue
		}

		jobResCount, err := dncJobResds.DncJobResultCountWithContactListId(list.ID)
		if err != nil {
			continue
		}

		if jobResCount < jobsCount {
			continue
		}

		limit := 5
		_, total, err := dncJobResds.DncJobResultWithContactListId(list.ID, 1, limit)

		if err != nil {
			logrus.Errorln(fmt.Sprintf("DncJobResultCountWithContactListId =>  contact list Id : %s :: Error :%s ::", list.ID.Hex(), err.Error()))
			continue
		}

		pages := int(math.Ceil(float64(total) / float64(limit)))
		for page := 1; page <= pages; page++ {
			res, _, err := dncJobResds.DncJobResultWithContactListId(list.ID, page, limit)
			if err != nil {
				contactListds.UpdateStatus(list.ID, model.ContactListStatusError)
				break
			}

			err = djrs.SaveDNCJobResultsToContactList(list.ID, res)
			if err != nil {
				list.Status = model.ContactListStatusError
				if err := contactListds.SaveContactList(&list); err != nil {
					logrus.Errorln(fmt.Sprintf("SaveContactList =>  contact list Id : %s :: Error :%s ::", list.ID.Hex(), err.Error()))
					// TODO: log error here
				}

				contactListds.UpdateStatus(list.ID, model.ContactListStatusError)
				break
			}
		}

		contactListds.UpdateDNCScrubStatus(list.ID, model.DNCScrubProcessStatusResponseReceived)
		contactListds.UpdateStatus(list.ID, model.ContactListStatusActive)
	}
}


func (djrs *DNCJobResultService) SaveDNCJobResultsToContactList(clid bson.ObjectId, results *model.DNCJobsResultList) error {
	// get new instance of contact list datasource
	contactListds := djrs.contactListDataSource()
	defer contactListds.Session.Close()

	// get new instance of contact group datasource
	cgds := djrs.contactGroupDataSource()
	defer cgds.Session.Close()

	for _, res := range *results {
		cl, err := contactListds.ContactListWithId(clid)
		if err != nil {
			return err
		}
		icg := &model.ContactGroup{
			ID:            bson.NewObjectId(),
			ContactListId: clid,
		}

		cl.ContactGroups = append(cl.ContactGroups, model.ContactGroupRef{ContactGroupId: icg.ID})

		var nums model.NumberList
		for _, i := range res.Clean {
			nt := djrs.ClassifyNumberType(i)
			tz := djrs.ClassifyNumberTimeZone(i)

			numTypeStr := ""
			// will eventually deprecate, its here because of design pattern mistake
			switch nt {
			case model.NumberTypeLandLine:
				numTypeStr = "Lindlinenumber"
				cl.LandLine += 1
			case model.NumberTypeCellPhone:
				numTypeStr = "cellnumber"
				cl.CellPhone += 1
			}

			nums = append(nums, model.Number{
				ID:            bson.NewObjectId(),
				Number:        i,
				TimeZone:      tz,
				NumberTypeStr: numTypeStr,
				NumberType:    nt,
			})
		}

		icg.Numbers = nums
		cl.DNCNumbers += int64(len(res.Matched))
		cl.NumberCount += int64(len(res.Clean))
		if err := cgds.SaveContactGroup(icg); err != nil {
			return nil
		}

		if err := contactListds.SaveContactList(cl); err != nil {
			return err
		}
	}

	return nil
}

func (djrs *DNCJobResultService) ProcessJobResultsForTTSList() {
	dncJobds := djrs.dncJobsDataSource()
	defer dncJobds.Session.Close()

	ttsListds := djrs.ttsListDataSource()
	defer ttsListds.Session.Close()

	dncJobResds := djrs.dncJobResultDataSource()
	defer dncJobResds.Session.Close()

	lists, _, err := ttsListds.GetListsWithDNCScrubStatus(model.DNCScrubProcessStatusAllBatchesSent, 1, 50)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("TTSListGetListWithDNCScrubStatus => Error :%s ::", err.Error()))
		return
	}

	for _, list := range *lists {
		jobsCount, err := dncJobds.DNCJobCountWithTTSId(list.ID)
		if err != nil {
			continue
		}

		jobResCount, err := dncJobResds.DNCJobResultCountWithTTSListId(list.ID)
		if err != nil {
			continue
		}

		if jobResCount < jobsCount {
			continue
		}

		limit := 5
		_, total, err := dncJobResds.DncJobResultWithttsListId(list.ID, 1, limit)

		if err != nil {
			logrus.Errorln(fmt.Sprintf("DncJobResultCountWithContactListId =>  contact list Id : %s :: Error :%s ::", list.ID.Hex(), err.Error()))
			continue
		}

		pages := int(math.Ceil(float64(total) / float64(limit)))
		for page := 1; page <= pages; page++ {
			res, _, err := dncJobResds.DncJobResultWithttsListId(list.ID, page, limit)
			if err != nil {
				ttsListds.UpdateStatus(list.ID, model.ContactListStatusError)
				break
			}

			err = djrs.SaveDNCJobResultsToTTSList(list.ID, res)
			if err != nil {
				list.Status = model.ContactListStatusError
				if err := ttsListds.SaveTTSList(list); err != nil {
					logrus.Errorln(fmt.Sprintf("SaveContactList =>  contact list Id : %s :: Error :%s ::", list.ID.Hex(), err.Error()))
					// TODO: log error here
				}

				ttsListds.UpdateStatus(list.ID, model.ContactListStatusError)
				break
			}
		}

		ttsListds.UpdateDNCScrubStatus(list.ID, model.DNCScrubProcessStatusResponseReceived)
		ttsListds.UpdateStatus(list.ID, model.ContactListStatusActive)
	}
}

func (djrs *DNCJobResultService) SaveDNCJobResultsToTTSList(clid bson.ObjectId, results *model.DNCJobsResultList) error {
	ttsListds := djrs.ttsListDataSource()
	defer ttsListds.Session.Close()

	ttsGroupds := djrs.ttsGroupDatasource()
	defer ttsGroupds.Session.Close()

	ttsMetads := djrs.ttsmetaDataSource()
	defer ttsMetads.Session.Close()

	for _, res := range *results {
		cl, err := ttsListds.TTSListWithId(clid)
		if err != nil {
			return err
		}
		icg := &model.TTSListContactGroup{
			ID:           bson.NewObjectId(),
			TTSListId:    clid,
			Numbers:      nil,
			CreateDate:   time.Time{},
			UpdateDate:   time.Time{},
			TotalNumbers: 0,
		}

		cl.ContactGroups = append(cl.ContactGroups, model.ContactGroupRef{ContactGroupId: icg.ID})

		var nums model.TTSNumberList
		for _, i := range res.Clean {
			nt := djrs.ClassifyNumberType(i)
			tz := djrs.ClassifyNumberTimeZone(i)

			numTypeStr := ""
			// will eventually deprecate, its here because of design pattern mistake
			switch nt {
			case model.NumberTypeLandLine:
				numTypeStr = "Lindlinenumber"
				cl.LandLine += 1
			case model.NumberTypeCellPhone:
				numTypeStr = "cellnumber"
				cl.CellPhone += 1
			}

			fd, err := ttsMetads.GetByNumber(i)
			if err != nil {
				continue
			}

			nums = append(nums, model.TTSNumber{
				ID:            bson.NewObjectId(),
				Number:        i,
				TimeZone:      tz,
				NumberTypeStr: numTypeStr,
				NumberType:    nt,
				FieldData:     fd.FieldData,
			})
		}

		icg.Numbers = nums
		cl.DNCNumbers += int64(len(res.Matched))
		cl.NumberCount += int64(len(res.Clean))
		if err := ttsGroupds.SaveContactGroup(icg); err != nil {
			return nil
		}

		if err := ttsListds.SaveTTSList(*cl); err != nil {
			return err
		}
	}

	return nil
}

func (djrs *DNCJobResultService) ClassifyNumberType(number string) model.NumberType {
	if len(number) < 10 {
		return model.NumberTypeUnknown
	}
	num := cmlutils.ExtractNumberFromUSNumber(number)
	prefixSix := num[0:6]

	if djrs.WirelessNumber(prefixSix) {
		return model.NumberTypeCellPhone
	} else if djrs.LandLineNumber(prefixSix) {
		return model.NumberTypeLandLine
	} else {
		return model.NumberTypeUnknown
	}
}

func (djrs *DNCJobResultService) WirelessNumber(number string) bool {
	s := djrs.RedisClient.HExists("cellnumber", number)
	str := s.String()
	if strings.Contains(str, "true") {
		return true
	} else {
		return false
	}
}

func (djrs *DNCJobResultService)LandLineNumber(number string) bool {
	s := djrs.RedisClient.HExists("landlinenumber", number)
	str := s.String()
	if strings.Contains(str, "true") {
		return true
	} else {
		return false
	}
}

func (djrs *DNCJobResultService) ClassifyNumberTimeZone(number string) string {
	if len(number) < 10 {
		return ""
	}
	num := cmlutils.ExtractNumberFromUSNumber(number)
	prefixThree := num[0:3]
	return djrs.Timezone(prefixThree)
}

func (djrs *DNCJobResultService) Timezone(number string) string {
	s := djrs.RedisClient.HGet("timezone", number)
	str := s.String()
	timezone := strings.Split(str, ":")
	tz := strings.TrimSpace(timezone[1])
	logrus.Infoln(tz)
	return tz
}
