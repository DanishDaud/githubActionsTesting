package service

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"gopkg.in/mgo.v2/bson"
)

var dncJobCronProcessingIndicator = make(chan int, 1)

func init() {
	// initial refill
	dncJobCronProcessingIndicator <- 1
}

type DNCJobService struct {
	Service
	TCPAService *TCPAService
}

func (djs *DNCJobService) dncJobsDataSource() *datasource.DNCJobsDataSource {
	return &datasource.DNCJobsDataSource{DataSource: datasource.DataSource{Session: djs.Session.Copy()}}
}

func (djs *DNCJobService) dncJobResultDataSource() *datasource.DNCJobsResultDataSource {
	return &datasource.DNCJobsResultDataSource{DataSource: datasource.DataSource{Session: djs.Session.Copy()}}
}

func (djs *DNCJobService) SaveDncJobObject(dncjobs *model.DNCJobs) error {
	// get new instance of dnc job datasource
	dncJobds := djs.dncJobsDataSource()
	defer dncJobds.Session.Close()

	return dncJobds.SaveDNCJob(dncjobs)
}
func (djs *DNCJobService) DeleteDncJobObject(dncjobs *model.DNCJobs) error {
	// get new instance of dnc job datasource
	dncJobds := djs.dncJobsDataSource()
	defer dncJobds.Session.Close()
	// get new instance of contact group datasource
	return dncJobds.DeleteDncJOB(dncjobs)
}

func (djs *DNCJobService) DncjobObjectContactlistId(objectID string, page int, limit int) (*model.DNCJobsList, int, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, 0, errors.New("object id is not valid")
	}

	dncjobObjectId := bson.ObjectIdHex(objectID)

	// get new instance of contact list datasource
	dncJobds := djs.dncJobsDataSource()
	defer dncJobds.Session.Close()

	return dncJobds.DNCObjectWithContactlistId(dncjobObjectId, page, limit)
}
func (djs *DNCJobService) DncjobObjectContactJobId(objectID string) (*model.DNCJobs, error) {

	// get new instance of contact list datasource
	dncJobds := djs.dncJobsDataSource()
	defer dncJobds.Session.Close()

	return dncJobds.DNCObjectWithJobId(objectID)
}

func (djs *DNCJobService) DncjobObjectWithId(objectID string) (*model.DNCJobs, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, errors.New("object id is not valid")
	}

	dncjobObjectId := bson.ObjectIdHex(objectID)

	// get new instance of contact list datasource
	dncJobds := djs.dncJobsDataSource()
	defer dncJobds.Session.Close()

	return dncJobds.DNCJobWithId(dncjobObjectId)
}

func (djs *DNCJobService) JobCountWithContactListId(objectID string) (int, error) {
	if !bson.IsObjectIdHex(objectID) {
		return 0, errors.New("object id is not valid")
	}

	contactlistId := bson.ObjectIdHex(objectID)
	contactListDataSource := djs.dncJobsDataSource()
	defer contactListDataSource.Session.Close()

	return contactListDataSource.DNCJobCountWithContactListId(contactlistId)
}

func (djs *DNCJobService) DNCScrubJobCron() {
	select {
	case temp := <-dncJobCronProcessingIndicator:
		logrus.Debugln("DNCJobResultCron Started ", temp)
		djs.ProcessDNCScrubJobCron()
		dncJobCronProcessingIndicator <- 1
		logrus.Debugln("DNCJobResultCron Ended ", temp)
	default:
		logrus.Debugln("Processing channel occupied")
	}
}

func (djs *DNCJobService) ProcessDNCScrubJobCron() {
	// get new instance of dnc job datasource
	dncJobds := djs.dncJobsDataSource()
	defer dncJobds.Session.Close()

	// get new instance of dnc job datasource
	dncJobResds := djs.dncJobResultDataSource()
	defer dncJobResds.Session.Close()

	jobs, err := dncJobds.DNCObjectsWithStatus(model.DNCScrubJobTypeProcessing)

	if err != nil {
		logrus.Errorln(fmt.Sprintf("TTSDncObjectWithStatus =>  Error :%s ::", err.Error()))
		return
	}
	m := make(map[string]interface{})

	for _, job := range *jobs {
		// steps
		// - check if job is completed, else ignore
		// - update job status to completed in database
		// - save job result in database
		resMap, err := djs.TCPAService.CheckTCPAJobStatus(job.JobId, m)
		if err != nil {
			logrus.Errorln(fmt.Sprintf("CheckTCPAJobStatus => Job Id => :%s :: Error :%s ::", job.JobId, err.Error()))
			// TODO: log this error here
			continue
		}

		var returned = false
		var match []string
		var clean []string

		if data, ok := resMap["match"]; ok {
			numMap, ok := data.(map[string]interface{})
			if ok {
				match = extractKeys(numMap)
			}
			returned = true
		}

		if data, ok := resMap["clean"]; ok {
			numMap, ok := data.(map[string]interface{})
			if ok {
				clean = extractKeys(numMap)
			}
			returned = true
		}

		if returned {
			// save job result
			err := dncJobResds.SaveDncJobResult(&model.DNCJobResult{
				Matched:       match,
				Clean:         clean,
				DncJobID:      job.ID,
				ContactListId: job.ContactListId,
				TTSListID:     job.TTSListID,
				JobId:         job.JobId,
			})
			if err != nil {
				logrus.Errorln("SaveDncJobResult====> TTS List Id :%s :: Job Id => :%s :: Error :%s ::", job.ContactListId, job.JobId, err.Error())
				continue
			}

			// job completed
			job.Status = model.DNCScrubJobTypeCompleted
			if err := dncJobds.SaveDNCJob(&job); err != nil {
				logrus.Errorln("SaveDNCJob====> Contact List Id :%s :: Job Id => :%s :: Error :%s ::", job.ContactListId, job.JobId, err.Error())
				// TODO: log error here
				continue
			}
		}
	}
}
