package service

import "C"
import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tealeg/xlsx"
	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlconstants"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

type ProcessTTSListBatchOptions struct {
	TTSList  *model.TTSList
	Batch    model.TTSNumberList
	UserInfo *model.User
}

type TTSListService struct {
	S3Service          *S3Service
	ContactListService ContactListService
	CustomFieldService CustomFieldService
	TTSGroup           TTSGroupService
	UserService        UserService
	TCPAService        TCPAService
	Service
}

func (smc *TTSListService) Import(obj model.ImportData) (*model.TTSList, error) {
	fileId := bson.ObjectIdHex(obj.FileID)
	fileInfo, err := smc.GetTTSFileInformation(fileId)

	if err != nil {
		return nil, err
	}

	uuid, _ := cmlutils.GetUUID()
	destinationPath := cmlconstants.TempDestinationPath + uuid + fileInfo.FileName
	url := cmlutils.S3FullPath() + fileInfo.S3Path
	err = smc.S3Service.DownloadFile(destinationPath, url)
	if err != nil {
		return nil, err
	}
	userInfo, errUser := smc.UserService.UserObject(obj.UserID)
	if errUser != nil {
		return nil, err
	}

	// features to provide
	// save contact list
	cl := &model.TTSList{}
	cl.ID = bson.NewObjectId()

	cl.FileName = fileInfo.FileName
	cl.Name = fileInfo.FileName
	cl.RemoveDup = obj.RemoveDup
	cl.Shuffle = obj.Shuffle
	cl.ScrubLandLine = obj.ScrubLandLine
	cl.ScrubCellPhone = obj.ScrubCellPhone
	cl.AreaCode = false
	cl.ScrubDNC = obj.ScrubDNC
	cl.FileS3Path = fileInfo.S3Path
	cl.UserID = fileInfo.UserID
	cl.Status = model.ContactListStatusProcessing
	cl.CreateDate = time.Now()

	if err := smc.SaveTTSList(cl); err != nil {
		logrus.Errorln("Error Save Contact list")
		logrus.Errorln(cl)
		return nil, err
	}

	go func() {
		smc.ProcessTTSFile(obj, destinationPath, *cl, userInfo)
	}()

	return cl, nil
}

func (smc *TTSListService) SaveTTSList(cl *model.TTSList) error {
	// get new instance of user data source
	sds := smc.ttsDataSource()
	defer sds.Session.Close()

	// get user object from database
	return sds.SaveTTSList(*cl)

}

func (smc *TTSListService) TTSList(cl *model.CustomFilter, all bool) (*[]model.TTSList, int, error) {
	// get new instance of user data source
	sds := smc.ttsDataSource()
	defer sds.Session.Close()

	// get user object from database
	return sds.List(*cl, all)

}

func (smc *TTSListService) GetTTSList(userid bson.ObjectId, listid bson.ObjectId) (*model.TTSList, error) {
	// get new instance of user data source
	sds := smc.ttsDataSource()
	defer sds.Session.Close()

	// get user object from database
	return sds.GET(userid, listid)

}

func (smc *TTSListService) DeleteList(listid bson.ObjectId) error {
	// get new instance of user data source
	sds := smc.ttsDataSource()
	defer sds.Session.Close()

	// get user object from database
	return sds.Delete(listid)
}

func (smc *TTSListService) ProcessTTSFile(meta model.ImportData, dst string, ttsList model.TTSList, user *model.User) error {
	var dupMap = make(map[string]bool)
	var batch []model.TTSNumber
	var cgRefs model.ContactGroupRefs

	const (
		Imported = iota
		Invalid
		Duplicate
		CellPhone
		LandLine
		Unknown
		ListTotal
	)
	var statMap = map[int8]int64{
		Imported:  0,
		Invalid:   0,
		Duplicate: 0,
		CellPhone: 0,
		LandLine:  0,
		ListTotal: 0,
	}
	xlFile, err := xlsx.OpenFile(dst)
	if err != nil {
		return err
	}

	if meta.Data.Number == -1 {
		logrus.Errorln("number Column is missing ::::Contact ID %s ::::File ID", ttsList.ID, meta.FileID)
		ttsList.Status = model.ContactListStatusError
		ttsList.Err = "missing phone column"
		err := smc.SaveTTSList(&ttsList)
		if err != nil {
			logrus.Errorln("Error :::SaveTTSList :::TTSListId %s ::::Error %s", ttsList.ID, err.Error())

		}
		cmlutils.DeleteFile(dst)
		return nil
	}

	var fieldsData = smc.generateField(meta)

	// First cycle the sheet, recycle the row, and finally the cell, print cell.String()
	for _, sheet := range xlFile.Sheets {
		if smc.colvalidation(model.DataColumn{NameCol: meta.Data.FirstName,
			LastNameCol: meta.Data.LastName,
			NumCol:      meta.Data.Number,
			EmailCol:    meta.Data.Email,
			AddressCol:  meta.Data.Address},
			len(sheet.Cols)) {
			logrus.Errorln("Invalid Data Column Select ::::Contact ID %s ::::File ID", ttsList.ID, meta.FileID)
			ttsList.Status = model.ContactListStatusError
			ttsList.Err = "invalid column data"
			err := smc.SaveTTSList(&ttsList)
			if err != nil {
				logrus.Errorln("Error :::SaveTTSList :::TTSListId %s ::::Error %s", ttsList.ID, err.Error())
				break
			}
			return nil
		}

		skipFirst := false
		for _, row := range sheet.Rows {
			// skip first row as it contains header
			if !skipFirst {
				skipFirst = true
				continue
			}

			if len(row.Cells) == 0 || !(meta.Data.Number > -1 && len(row.Cells) >= meta.Data.Number) {
				continue
			}

			logrus.Errorf("Number : %d :: Length : %d", meta.Data.Number, len(row.Cells))
			numbercell := row.Cells[meta.Data.Number]

			statMap[Imported] += 1
			number := cmlutils.SimplifyPhoneNumber(numbercell.Value)
			if !cmlutils.IsUSNumber(number) {
				statMap[Invalid] += 1
				continue
			}
			numInfo := smc.ContactListService.CreateNumberInfo(numbercell.Value)

			if ttsList.RemoveDup {
				_, ok := dupMap[numInfo.Number]
				if ok {
					// duplicate
					statMap[Duplicate] += 1
					continue
				} else {
					dupMap[numInfo.Number] = true
				}
			}
			ttsNumberInfo := model.TTSNumber{
				ID:            numInfo.ID,
				Number:        numInfo.Number,
				NumberType:    numInfo.NumberType,
				TimeZone:      numInfo.TimeZone,
				NumberTypeStr: numInfo.NumberTypeStr,
				FieldData:     smc.generateFieldData(meta, row),
			}

			if ttsList.ScrubLandLine && numInfo.NumberType == model.NumberTypeLandLine {
				// drop cell phone numbers
				statMap[Invalid] += 1
				continue
			}

			if ttsList.ScrubCellPhone && numInfo.NumberType == model.NumberTypeCellPhone {
				// drop landline numbers
				statMap[Invalid] += 1
				continue
			}

			if numInfo.NumberType == model.NumberTypeCellPhone {
				statMap[CellPhone] += 1
			} else if numInfo.NumberType == model.NumberTypeLandLine {
				statMap[LandLine] += 1
			} else {
				statMap[Unknown] += 1
			}



			batch = append(batch, ttsNumberInfo)

			if len(batch) == cmlconstants.ConfigContactListBatchSize {
				cg, err := smc.ProcessTTSListBatch(&ProcessTTSListBatchOptions{
					TTSList:  &ttsList,
					Batch:    batch,
					UserInfo: user,
				})
				if err != nil {
					cmlutils.DeleteFile(dst)
					logrus.Errorln(fmt.Sprintf("processContactListBatch => contact list Id : %s :: Error :%s ::", ttsList.ID.Hex(), err.Error()))
					return err
				}

				if cg != nil {
					statMap[ListTotal] += int64(len(batch))
					cgRef := model.ContactGroupRef{ContactGroupId: cg.ID}
					cgRefs = append(cgRefs, cgRef)
				}
				batch = model.TTSNumberList{}
			}
		}

		break
	}

	if len(batch) > 0 {
		cg, err := smc.ProcessTTSListBatch(&ProcessTTSListBatchOptions{
			TTSList:  &ttsList,
			Batch:    batch,
			UserInfo: user,
		})
		if err != nil {
			logrus.Errorln(fmt.Sprintf("processContactListBatch => contact list Id : %s :: Error :%s ::", ttsList.ID.Hex(), err.Error()))
			cmlutils.DeleteFile(dst)
			return err
		}

		if cg != nil {
			statMap[ListTotal] += int64(len(batch))
			cgRef := model.ContactGroupRef{ContactGroupId: cg.ID}
			cgRefs = append(cgRefs, cgRef)
		}
		batch = model.TTSNumberList{}
	}

	ttsList.ContactGroups = cgRefs
	ttsList.Imported = statMap[Imported]
	ttsList.NumberCount = statMap[ListTotal]
	ttsList.Invalid = statMap[Invalid]
	ttsList.CellPhone = statMap[CellPhone]
	ttsList.Duplicate = statMap[Duplicate]
	ttsList.LandLine = statMap[LandLine]
	ttsList.Unknown = statMap[Unknown]
	ttsList.DNCNumbers = 0
	ttsList.Status = model.ContactListStatusActive
	ttsList.FieldData = fieldsData

	if ttsList.ScrubDNC {
		ttsList.Status = model.ContactListStatusProcessing
		ttsList.DNCScrubStatus = model.DNCScrubProcessStatusAllBatchesSent
		ttsList.LandLine = 0
		ttsList.CellPhone = 0
	} else {
		ttsList.Status = model.ContactListStatusActive
	}

	if err := smc.SaveTTSList(&ttsList); err != nil {
		logrus.Errorln(fmt.Sprintf("saveContactList => contact list Id : %s :: Error :%s ::", ttsList.ID.Hex(), err.Error()))
		cmlutils.DeleteFile(dst)
		return err
	}
	cmlutils.DeleteFile(dst)
	return nil
}

func (smc *TTSListService) generateField(meta model.ImportData) (data []model.Fields) {
	if meta.Data.Number != -1 {
		data = append(data, model.Fields{
			Name:    string(model.TTSDefaultFieldTypeNumber),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeNumber),
		})
	}

	if meta.Data.FirstName != -1 {
		data = append(data, model.Fields{
			Name:    string(model.TTSDefaultFieldTypeFirstName),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeFirstName),
		})
	}

	if meta.Data.LastName != -1 {
		data = append(data, model.Fields{
			Name:    string(model.TTSDefaultFieldTypeLastName),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeLastName),
		})
	}

	if meta.Data.Email != -1 {
		data = append(data, model.Fields{
			Name:    string(model.TTSDefaultFieldTypeEmail),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeEmail),
		})
	}

	if meta.Data.Address != -1 {
		data = append(data, model.Fields{
			Name:    string(model.TTSDefaultFieldTypeAddress),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeAddress),
		})
	}

	if meta.Data.Location != -1 {
		data = append(data, model.Fields{
			Name:    string(model.TTSDefaultFieldTypeLocation),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeLocation),
		})
	}

	if meta.Data.City != -1 {
		data = append(data, model.Fields{
			Name:    string(model.TTSDefaultFieldTypeCity),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeCity),
		})
	}

	if meta.Data.State != -1 {
		data = append(data, model.Fields{
			Name:    string(model.TTSDefaultFieldTypeState),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeState),
		})
	}

	if meta.Data.Country != -1 {
		data = append(data, model.Fields{
			Name:    string(model.TTSDefaultFieldTypeCountry),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeCountry),
		})
	}

	for _, i := range meta.Data.Customfields {
		name, key := i.NameAndKey()
		data = append(data, model.Fields{
			Name:    name,
			NameKey: key,
		})
	}

	return
}

func (smc *TTSListService) generateFieldData(meta model.ImportData, row *xlsx.Row) (data []model.TTSFieldData) {
	if meta.Data.Number > -1 && len(row.Cells) >= meta.Data.Number {
		record := model.TTSFieldData{
			Name:    string(model.TTSDefaultFieldTypeNumber),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeNumber),
			Data:    row.Cells[meta.Data.Number].Value,
		}
		data = append(data, record)
	}

	if meta.Data.FirstName > -1 && len(row.Cells) >= meta.Data.FirstName {
		record := model.TTSFieldData{
			Name:    string(model.TTSDefaultFieldTypeFirstName),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeFirstName),
			Data:    row.Cells[meta.Data.FirstName].Value,
		}
		data = append(data, record)
	}

	if meta.Data.LastName > -1 && len(row.Cells) >= meta.Data.LastName {
		record := model.TTSFieldData{
			Name:    string(model.TTSDefaultFieldTypeLastName),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeLastName),
			Data:    row.Cells[meta.Data.LastName].Value,
		}
		data = append(data, record)
	}

	if meta.Data.Address > -1 && len(row.Cells) >= meta.Data.Address {
		record := model.TTSFieldData{
			Name:    string(model.TTSDefaultFieldTypeAddress),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeAddress),
			Data:    row.Cells[meta.Data.Address].Value,
		}
		data = append(data, record)
	}

	if meta.Data.Email > -1 && len(row.Cells) >= meta.Data.Email {
		record := model.TTSFieldData{
			Name:    string(model.TTSDefaultFieldTypeEmail),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeEmail),
			Data:    row.Cells[meta.Data.Email].Value,
		}
		data = append(data, record)
	}

	if meta.Data.Location > -1 && len(row.Cells) >= meta.Data.Location {
		if len(row.Cells) >= meta.Data.Location {
			record := model.TTSFieldData{
				Name:    string(model.TTSDefaultFieldTypeLocation),
				NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeLocation),
				Data:    row.Cells[meta.Data.Location].Value,
			}
			data = append(data, record)
		}
	}

	if meta.Data.Country > -1 && len(row.Cells) >= meta.Data.Country {
		record := model.TTSFieldData{
			Name:    string(model.TTSDefaultFieldTypeCountry),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeCountry),
			Data:    row.Cells[meta.Data.Country].Value,
		}
		data = append(data, record)
	}

	if meta.Data.City > -1 && len(row.Cells) >= meta.Data.City {
		record := model.TTSFieldData{
			Name:    string(model.TTSDefaultFieldTypeCity),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeCity),
			Data:    row.Cells[meta.Data.City].Value,
		}
		data = append(data, record)
	}

	if meta.Data.State > -1 && len(row.Cells) >= meta.Data.State {
		record := model.TTSFieldData{
			Name:    string(model.TTSDefaultFieldTypeState),
			NameKey: model.GetTTSDefaultFieldValueForKey(model.TTSDefaultFieldTypeState),
			Data:    row.Cells[meta.Data.State].Value,
		}
		data = append(data, record)
	}

	for _, i := range meta.Data.Customfields {
		name, key := i.NameAndKey()

		record := model.TTSFieldData{
			Name:    name,
			NameKey: key,
			Data:    row.Cells[i.Column].Value,
		}
		data = append(data, record)
	}

	return
}

func (smc *TTSListService) find(slice []model.Fields, val string) (int, bool) {
	for i, item := range slice {
		if item.Name == val {
			return i, true
		}
	}
	return -1, false
}

func (smc *TTSListService) fields(row *xlsx.Row, col int) model.Fields {

	field := model.Fields{}
	name := row.Cells[col].Value
	key := strings.ToLower(name)
	key = strings.Replace(key, " ", "_", -1)
	field.Name = name
	field.NameKey = key

	return field
}

func (smc *TTSListService) colvalidation(datacol model.DataColumn, colsize int) bool {

	if (datacol.AddressCol >= colsize) || (datacol.EmailCol >= colsize) || (datacol.NumCol >= colsize) || (datacol.NameCol >= colsize) || (datacol.LastNameCol >= colsize) {
		return true
	}
	return false
}

func (smc *TTSListService) ProcessTTSListBatch(op *ProcessTTSListBatchOptions) (*model.TTSListContactGroup, error) {
	// get new instance of contact group datasource
	sds := smc.ttsDataSource()
	defer sds.Session.Close()

	if op.TTSList.ScrubDNC {
		if err := smc.DNCScrubTTSList(op); err != nil {
			logrus.Errorln(fmt.Sprintf("Error  DNCScrubContactList  => contact list Id : %s ::user Id : %s :: Error :%s ::", op.TTSList.ID.Hex(), op.UserInfo.ID.Hex(), err.Error()))
			return nil, err
		}
		return nil, nil
	} else {
		cg := model.TTSListContactGroup{}
		cg.ID = bson.NewObjectId()
		cg.TTSListId = op.TTSList.ID
		cg.Numbers = op.Batch
		err := smc.TTSGroup.SaveContactGroup(&cg)
		if err != nil {
			logrus.Errorln(fmt.Sprintf("saveContactGroup => contact list Id : %s :: contact group Id : %s :: Error :%s ::", op.TTSList.ID.Hex(), cg.ID, err.Error()))
			return nil, err
		}
		return &cg, nil
	}

}

func (smc *TTSListService) SaveTTSFile(obj model.TTSFile) (*model.TTSFile, error) {
	// get new instance of user data source
	sds := smc.ttsDataSource()
	defer sds.Session.Close()

	// get user object from database
	return sds.SaveFile(obj)
}

func (smc *TTSListService) GetTTSFileInformation(objectId bson.ObjectId) (*model.TTSFile, error) {
	// get new instance of user data source
	sds := smc.ttsDataSource()
	defer sds.Session.Close()

	// get user object from database
	return sds.GetFile(objectId)
}

func (smc *TTSListService) ttsDataSource() *datasource.TTSListDataSource {
	return &datasource.TTSListDataSource{DataSource: datasource.DataSource{Session: smc.Session.Copy()}}
}

func (smc *TTSListService) ttsmetaDataSource() *datasource.TTSMetaDataSource {
	return &datasource.TTSMetaDataSource{DataSource: datasource.DataSource{Session: smc.Session.Copy()}}
}

// return instance if contact group  datasource
// every time a new instance would be created
func (cgs *TTSGroupService) ttsGroupDatasource() *datasource.TTSGroupDataSource {
	return &datasource.TTSGroupDataSource{DataSource: datasource.DataSource{Session: cgs.Session.Copy()}}
}

func (cls *TTSListService) DNCScrubTTSList(op *ProcessTTSListBatchOptions) error {
	// get new instance of dnc job datasource
	dncJobds := cls.ttsDataSource()

	defer dncJobds.Session.Close()

	var chunkSize = cmlconstants.ConfigContactListBatchSize
	var errs []error
	var divided = splittts(op.Batch, chunkSize)

	m := make(map[string]interface{})
	m["ttsListId"] = op.TTSList.ID.Hex()
	m["userId"] = op.UserInfo.ID.Hex()
	var wg sync.WaitGroup
	// send api calls
	for _, chunk := range divided {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var numbersInChunk []string
			for _, num := range chunk {
				numbersInChunk = append(numbersInChunk, num.Number)
			}

			err := cls.SaveTTSNumberMeta(op.TTSList, chunk)
			if err != nil {
				errs = append(errs, err)
				return
			}

			jobId, err := cls.TCPAService.IsNumberInTCPA(numbersInChunk, m)
			if err != nil {
				logrus.Errorln(fmt.Sprintf("IsNumberInTCPA  => contact list Id : %s ::user Id : %s :: Error :%s ::", op.TTSList.ID.Hex(), op.UserInfo.ID.Hex(), err.Error()))
				errs = append(errs, err)
				return
			}

			job := model.DNCJobs{
				Status:    model.DNCScrubJobTypeProcessing,
				TTSListID: &op.TTSList.ID,
				JobId:     jobId,
			}

			err = dncJobds.SaveDncJob(&job)
			if err != nil {
				logrus.Errorln(fmt.Sprintf("SaveDNCJob => contact list Id : %s ::user Id : %s  ::job Id : %s :: Error :%s ::", op.TTSList.ID.Hex(), op.UserInfo.ID.Hex(), jobId, err.Error()))
				errs = append(errs, err)
			}
		}()
	}

	wg.Wait()

	if len(errs) > 0 {
		logrus.Errorln(fmt.Sprintf("IsNumberInTCPA  => contact list Id : %s ::user Id : %s :: Error :%s ::", op.TTSList.ID.Hex(), op.UserInfo.ID.Hex(), errs[0].Error()))
		// TODO: compose proper error from all errors and return
		return errs[0]
	}

	return nil
}

func (cls *TTSListService) SaveTTSNumberMeta(ttsList *model.TTSList, numbers []model.TTSNumber) error {
	ttsMeta := cls.ttsmetaDataSource()
	var numberMetas model.TTSMetaList
	for _, num := range numbers {
		meta := model.TTSMetaNumber{
			TTSListID:     ttsList.ID,
			Number:        num.Number,
			TimeZone:      num.TimeZone,
			NumberTypeStr: num.NumberTypeStr,
			NumberType:    num.NumberType,
			FieldData:     num.FieldData,
			ID: num.ID,
		}

		numberMetas = append(numberMetas, meta)
	}
	return ttsMeta.BulkInsert(numberMetas)
}


