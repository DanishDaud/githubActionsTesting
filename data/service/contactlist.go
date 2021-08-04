package service

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlconstants"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlmessages"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

// Data Structure to represent contact list service
type ContactListService struct {
	Service
	RedisService
	S3Service   *S3Service
	TCPAService *TCPAService
	Fileservice FileService
}

type ContactListSaveOptions struct {
	Name            string
	FileName        string
	FilePath        string
	UserInfo        *model.User
	NumberColumn    int8
	TextColumn      int8
	HeaderPresent   bool
	ContactListType model.ContactListType
	RemoveDuplicate bool
	ShuffleNumber   bool
	ScrubLandline   bool
	ScrubCellPhone  bool
	AreaCodeDialing bool
	ScrubDNC        bool
	Randomize       bool
	File            *os.File
	Reader          *csv.Reader
}

type ProcessContactListOptions struct {
	ContactList   *model.ContactList
	UserInfo      *model.User
	FilePath      string
	NumberColumn  int8
	TextColumn    int8
	HeaderPresent bool
	File          *os.File
	Reader        *csv.Reader
}

type ProcessContactListBatchOptions struct {
	Writer      *csv.Writer
	ContactList *model.ContactList
	UserInfo    *model.User
	Batch       model.NumberList
}

func (cls *ContactListService) ContactListObjectWithId(objectID string) (*model.ContactList, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, errors.New("object id is not valid")
	}

	contactListObjectId := bson.ObjectIdHex(objectID)

	// get new instance of contact list datasource
	contactListDataSource := cls.contactListDataSource()

	defer contactListDataSource.Session.Close()

	return contactListDataSource.ContactListWithId(contactListObjectId)
}
func (cls *ContactListService) DeleteDNCNumber(contactList *model.ContactList, number string) error {
	// get new instance of contact list datasource
	contactListDataSource := cls.contactListDataSource()
	defer contactListDataSource.Session.Close()

	// get new instance of contact group datasource
	contactGroupDataSource := cls.contactGroupDataSource()
	defer contactGroupDataSource.Session.Close()

	// Delete all associated contact groups as well
	if err := contactGroupDataSource.DeleteDNCNumber(contactList.ID, number); err != nil {
		logrus.Infoln("Contact Groups delete failed => Contact list Id : " + contactList.ID.Hex())
		return err
	}
	return nil
	//	return contactListDataSource.DeleteContactList(contactList)
}
func (cls *ContactListService) DoNotContactListWithUserId(userId string) (model.ContactLists, error) {
	if !bson.IsObjectIdHex(userId) {
		return nil, errors.New("object id is not valid")
	}

	userObjectId := bson.ObjectIdHex(userId)

	// get new instance of contact list datasource
	contactListDataSource := cls.contactListDataSource()
	defer contactListDataSource.Session.Close()

	return contactListDataSource.DoNotContactListWithUserId(userObjectId)
}

func (cls *ContactListService) ContactListObjectWithIdAndUserId(objectID string, userId bson.ObjectId) (*model.ContactList, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, errors.New("object id is not valid")
	}

	//if !bson.IsObjectIdHex(userId) {
	//	return nil, errors.New("user id is not valid")
	//	}

	contactListObjectId := bson.ObjectIdHex(objectID)
	//	userObjectId := bson.ObjectIdHex(userId)

	// get new instance of contact list datasource
	contactListDataSource := cls.contactListDataSource()
	defer contactListDataSource.Session.Close()

	return contactListDataSource.ContactListWithIdAndUserId(contactListObjectId, userId)
}

func (cls *ContactListService) ContactListListingWithUserId(objectID string, page int, limit int, ctype int8, all bool) (*model.ContactLists, int, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, 0, errors.New("object id is not valid")
	}

	userObjectId := bson.ObjectIdHex(objectID)

	// get new instance of contact list datasource
	contactListDataSource := cls.contactListDataSource()
	defer contactListDataSource.Session.Close()

	return contactListDataSource.ContactListGetList(userObjectId, page, limit, ctype, all)
}

func (cls *ContactListService) SaveContactList(contactList *model.ContactList) error {
	// get new instance of sound file datasource
	contactListDataSource := cls.contactListDataSource()
	defer contactListDataSource.Session.Close()

	return contactListDataSource.SaveContactList(contactList)
}

func (cls *ContactListService) UpdateContactList(contactId string, s3filepath string, op *ContactListSaveOptions) (*model.ContactList, error) {

	// get new instance of contact list datasource
	clds := cls.contactListDataSource()
	defer clds.Session.Close()

	clid := bson.NewObjectId()

	// upload file to s3
	destinationPath := "/account/" + op.UserInfo.ID.Hex() + "/contactlist/" + clid.Hex() + "/"
	s3FilePath, err := cls.S3Service.Upload(destinationPath, op.FilePath)
	if err != nil {
		// TODO: file is not deleted
		return nil, err
	}

	list, err := cls.ContactListObjectWithIdAndUserId(contactId, op.UserInfo.ID)
	if err != nil {
		return nil, err
	}

	list.Status = model.ContactListStatusProcessing

	list = cls.ValidateContactListFlags(list)
	if err := clds.SaveContactList(list); err != nil {
		// delete file from s3
		cls.S3Service.DeleteS3Object(cmlutils.S3BucketName(), s3FilePath)
		logrus.Errorln("Error Save Contact list")
		logrus.Errorln(list)
		return nil, err
	}

	// contact list saved, not continue extraction of numbers in background mode
	go func() {
		err := cls.processUpdateContactList(&ProcessContactListOptions{
			ContactList:   list,
			UserInfo:      op.UserInfo,
			FilePath:      op.FilePath,
			NumberColumn:  op.NumberColumn,
			HeaderPresent: op.HeaderPresent,
		})

		if err != nil {
			logrus.Errorln(fmt.Sprintf("processContactList file failed => contact list Id : %s :: Error :%s ::", list.ID.Hex(), err.Error()))
			// TODO: Log error here
		}
	}()

	return list, nil

}
func (cls *ContactListService) SaveContactListNew(op *ContactListSaveOptions) (*model.ContactList, error) {
	// get new instance of contact list datasource
	clds := cls.contactListDataSource()
	defer clds.Session.Close()

	clid := bson.NewObjectId()

	// upload file to s3
	destinationPath := "/account/" + op.UserInfo.ID.Hex() + "/contactlist/" + clid.Hex() + "/"
	s3FilePath, err := cls.S3Service.Upload(destinationPath, op.FilePath)
	if err != nil {
		// TODO: file is not deleted
		return nil, err
	}

	// features to provide
	// save contact list
	cl := &model.ContactList{}
	cl.ID = clid
	cl.FileName = op.FileName
	cl.Name = op.Name
	cl.Type = op.ContactListType
	cl.RemoveDup = op.RemoveDuplicate
	cl.Shuffle = op.ShuffleNumber
	cl.ScrubLandLine = op.ScrubLandline
	cl.ScrubCellPhone = op.ScrubCellPhone
	cl.AreaCode = op.AreaCodeDialing
	cl.ScrubDNC = op.ScrubDNC
	cl.FileS3Path = s3FilePath
	cl.UserID = op.UserInfo.ID
	cl.Status = model.ContactListStatusProcessing

	cl = cls.ValidateContactListFlags(cl)

	if err := clds.SaveContactList(cl); err != nil {
		// delete file from s3
		cls.S3Service.DeleteS3Object(cmlutils.S3BucketName(), s3FilePath)
		logrus.Errorln("Error Save Contact list")
		logrus.Errorln(cl)
		return nil, err
	}

	// contact list saved, not continue extraction of numbers in background mode
	go func() {
		err := cls.ProcessContactList(&ProcessContactListOptions{
			ContactList:   cl,
			UserInfo:      op.UserInfo,
			FilePath:      op.FilePath,
			NumberColumn:  op.NumberColumn,
			TextColumn:    op.TextColumn,
			HeaderPresent: op.HeaderPresent,
		})

		if err != nil {
			logrus.Errorln(fmt.Sprintf("processContactList file failed => contact list Id : %s :: Error :%s ::", cl.ID.Hex(), err.Error()))
			// TODO: Log error here
		}
	}()

	return cl, nil
}

func (cls *ContactListService) ValidateContactListFlags(cl *model.ContactList) *model.ContactList {
	// => DNC
	// - All options false
	// => Default
	// Shuffle
	// Remove Duplicate
	// Scrub DNC
	// Scrub Landline
	// Scrub CellPhone
	// => CallerId
	// Area Code dialing
	// Randomize

	switch cl.Type {
	case model.ContactListTypeCallerGroups:
		cl.Shuffle = false
		cl.RemoveDup = false
		cl.ScrubDNC = false
		cl.ScrubLandLine = false
		cl.ScrubCellPhone = false
		break
	case model.ContactListTypeDNC:
		cl.Shuffle = false
		cl.RemoveDup = false
		cl.ScrubDNC = false
		cl.ScrubLandLine = false
		cl.ScrubCellPhone = false
		cl.Random = false
	}

	return cl
}

func (cls *ContactListService) ProcessContactList(op *ProcessContactListOptions) error {
	// read file
	inFile, inFileReader, err := cls.ReadFile(op.FilePath)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("open local file failed => contact list Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), err.Error()))
		return err
	}
	defer inFile.Close()

	//	defer op.File.Close()
	outFileName, err := cmlutils.RemoveNonAlphaNumeric(op.ContactList.Name)
	if err != nil {
		outFileName = "contacts"
	}

	outFilePath := fmt.Sprintf("%s%s.csv", cmlconstants.TempDestinationPath, outFileName)
	outFile, err := os.Create(outFilePath)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("create outfilepath  failed => contact list Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), err.Error()))
		return err
	}

	writer := csv.NewWriter(outFile)

	defer cmlutils.DeleteFile(op.FilePath)
	defer cmlutils.DeleteFile(outFilePath)

	headerRowTraversed := false

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
	var batch model.NumberList
	var cgRefs model.ContactGroupRefs
	var dupMap = make(map[string]bool)

	for {
		var text string
		record, err := inFileReader.Read()

		if err != nil {
			if err == io.EOF {
				logrus.Infoln("EOF reached")
				cg, err := cls.ProcessContactListBatch(&ProcessContactListBatchOptions{
					ContactList: op.ContactList,
					UserInfo:    op.UserInfo,
					Batch:       batch,
					Writer:      writer,
				})
				if err != nil {
					logrus.Errorln(fmt.Sprintf("processContactListBatch => contact list Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), err.Error()))
					return err
				}

				if cg != nil {
					statMap[ListTotal] += int64(len(batch))
					cgRef := model.ContactGroupRef{ContactGroupId: cg.ID}
					cgRefs = append(cgRefs, cgRef)
				}

				batch = model.NumberList{}
				break
			} else {
				logrus.Infoln("Error occurred")
			}
		} else {
			if op.HeaderPresent && !headerRowTraversed {
				headerRowTraversed = true
				continue
			}

			statMap[Imported] += 1

			// valid row found
			num, err := cls.extractNumberFromRow(record, op.NumberColumn)
			if err != nil {
				statMap[Invalid] += 1
				logrus.Errorln(fmt.Sprintf("extractNumberFromRow => contact list Id : %s :: Number : %s :: Num Col : %d", op.ContactList.ID.Hex(), record, op.NumberColumn))
				continue
			}
			if op.TextColumn >= 0 {
				text, err = cls.extractTextFromRow(record, op.TextColumn)
				if err != nil {
					statMap[Invalid] += 1
					logrus.Errorln(fmt.Sprintf("extractTextFromRow => contact list Id : %s :: Number : %s :: Text Col : %d", op.ContactList.ID.Hex(), record, op.TextColumn))
					continue
				}

				if text == "" {
					statMap[Invalid] += 1
					logrus.Errorln(fmt.Sprintf("extractTextFromRow => contact list Id : %s :: Number : %s :: Text Col : %d", op.ContactList.ID.Hex(), record, op.TextColumn))
					continue
				}
			}

			numInfo := cls.CreateNumberInfo(num)
			numInfo.Text = text
			if op.ContactList.Type == model.ContactListTypeDefault && !op.ContactList.RemoveDup {

			} else {
				_, ok := dupMap[numInfo.Number]
				if ok {
					// duplicate
					statMap[Duplicate] += 1
					continue
				} else {
					dupMap[numInfo.Number] = true
				}
			}

			if (op.ContactList.Type == model.ContactListTypeCallerGroups && !numInfo.IsValidCallerId()) ||
				(op.ContactList.Type == model.ContactListTypeDefault && !cmlutils.IsUSNumber(num)) ||
				(op.ContactList.Type == model.ContactListTypeDNC && !cmlutils.IsUSNumber(num)) {
				statMap[Invalid] += 1
				continue
			}

			if op.ContactList.ScrubCellPhone && numInfo.NumberType == model.NumberTypeCellPhone {
				// drop cell phone numbers
				statMap[Invalid] += 1
				continue
			}

			if op.ContactList.ScrubLandLine && numInfo.NumberType == model.NumberTypeLandLine {
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

			batch = append(batch, *numInfo)

			if len(batch) == cmlconstants.ConfigContactListBatchSize {
				cg, err := cls.ProcessContactListBatch(&ProcessContactListBatchOptions{
					ContactList: op.ContactList,
					UserInfo:    op.UserInfo,
					Batch:       batch,
					Writer:      writer,
				})
				if err != nil {
					logrus.Errorln(fmt.Sprintf("processContactListBatch => contact list Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), err.Error()))
					return err
				}

				if cg != nil {
					statMap[ListTotal] += int64(len(batch))
					cgRef := model.ContactGroupRef{ContactGroupId: cg.ID}
					cgRefs = append(cgRefs, cgRef)
				}
				batch = model.NumberList{}
			}
		}
	}

	op.ContactList.ContactGroups = cgRefs
	op.ContactList.Imported = statMap[Imported]
	op.ContactList.NumberCount = statMap[ListTotal]
	op.ContactList.Invalid = statMap[Invalid]
	op.ContactList.CellPhone = statMap[CellPhone]
	op.ContactList.Duplicate = statMap[Duplicate]
	op.ContactList.LandLine = statMap[LandLine]
	op.ContactList.Unknown = statMap[Unknown]
	op.ContactList.DNCNumbers = 0

	// save newly created file to s3
	s3FilePath, err := cls.uploadContactListFile(op.UserInfo.ID, op.ContactList.ID, outFilePath)
	if err == nil {
		op.ContactList.FileName = outFileName
		op.ContactList.FileS3Path = s3FilePath
		// TODO: delete previously uploaded file from s3
	}

	if op.ContactList.Type == model.ContactListTypeDefault && op.ContactList.ScrubDNC {
		op.ContactList.Status = model.ContactListStatusProcessing
		op.ContactList.DNCScrubStatus = model.DNCScrubProcessStatusAllBatchesSent
		op.ContactList.LandLine = 0
		op.ContactList.CellPhone = 0
	} else {
		op.ContactList.Status = model.ContactListStatusActive
	}

	if err := cls.SaveContactList(op.ContactList); err != nil {
		logrus.Errorln(fmt.Sprintf("saveContactList => contact list Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), err.Error()))
		return err
	}

	return nil
}

func (cls *ContactListService) processUpdateContactList(op *ProcessContactListOptions) error {
	// read file
	inFile, err := os.Open(op.FilePath)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("open local file failed => contact list Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), err.Error()))
		return err
	}
	defer inFile.Close()

	outFileName, err := cmlutils.RemoveNonAlphaNumeric(op.ContactList.Name)
	if err != nil {
		outFileName = "contacts"
	}

	outFilePath := fmt.Sprintf("%s%s.csv", cmlconstants.TempDestinationPath, outFileName)
	outFile, err := os.Create(outFilePath)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("create outfilepath  failed => contact list Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), err.Error()))
		return err
	}

	inFileReader := csv.NewReader(inFile)
	writer := csv.NewWriter(outFile)
	defer func() {
		writer.Flush()
		cmlutils.DeleteFile(op.FilePath)
		cmlutils.DeleteFile(outFilePath)
	}()

	headerRowTraversed := false

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

	var lastId model.ContactGroupRef
	if len(op.ContactList.ContactGroups) > 0 {
		lastId = op.ContactList.ContactGroups[len(op.ContactList.ContactGroups)-1]
	} else {
		createContactGroupObject(op.ContactList)
	}
	contactgroup, err := cls.contactGroupDataSource().ContactGroupWithId(lastId.ContactGroupId)
	if err != nil {
		contactgroup = createContactGroupObject(op.ContactList)
		//return nil, err
	}

	var rnumber = int(cmlconstants.ConfigContactListBatchSize - contactgroup.TotalNumbers)
	var batch model.NumberList
	var cgRefs model.ContactGroupRefs
	var dupMap = make(map[string]bool)
	cgRefs = append(cgRefs, op.ContactList.ContactGroups...)
	for {
		record, err := inFileReader.Read()
		if err != nil {
			if err == io.EOF {
				lastId = cgRefs[len(cgRefs)-1]
				contactgroup, err := cls.contactGroupDataSource().ContactGroupWithId(lastId.ContactGroupId)
				if err != nil {
					contactgroup = createContactGroupObject(op.ContactList)
					cgRef := model.ContactGroupRef{ContactGroupId: contactgroup.ID}
					cgRefs = append(cgRefs, cgRef)
				}
				contactgroup.Numbers = append(contactgroup.Numbers, batch...)
				if err := cls.contactGroupDataSource().SaveContactGroup(contactgroup); err != nil {
					return err
				}

				statMap[ListTotal] += int64(len(batch))
				logrus.Errorln("batch length", statMap[ListTotal])
				batch = model.NumberList{}
				break

			} else {
				logrus.Infoln("Error occurred")
			}
		} else {
			if op.HeaderPresent && !headerRowTraversed {
				headerRowTraversed = true
				continue
			}

			statMap[Imported] += 1

			// valid row found
			num, err := cls.extractNumberFromRow(record, op.NumberColumn)
			if err != nil {
				statMap[Invalid] += 1
				logrus.Errorln(fmt.Sprintf("extractNumberFromRow => contact list Id : %s :: Number : %s :: Num Col : %d", op.ContactList.ID.Hex(), record, op.NumberColumn))
				continue
			}
			numInfo := cls.CreateNumberInfo(num)

			_, ok := dupMap[numInfo.Number]
			if ok {
				// duplicate
				statMap[Duplicate] += 1
				continue
			} else {
				dupMap[numInfo.Number] = true
			}

			if (op.ContactList.Type == model.ContactListTypeCallerGroups && !numInfo.IsValidCallerId()) ||
				(op.ContactList.Type == model.ContactListTypeDefault && !cmlutils.IsUSNumber(num)) ||
				(op.ContactList.Type == model.ContactListTypeDNC && !cmlutils.IsUSNumber(num)) {
				statMap[Invalid] += 1
				continue
			}

			if op.ContactList.ScrubCellPhone && numInfo.NumberType == model.NumberTypeCellPhone {
				// drop cell phone numbers
				statMap[Invalid] += 1
				continue
			}

			if op.ContactList.ScrubLandLine && numInfo.NumberType == model.NumberTypeLandLine {
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

			batch = append(batch, *numInfo)

			if contactgroup.TotalNumbers < cmlconstants.ConfigContactListBatchSize {
				if len(batch) == rnumber {
					contactgroup.Numbers = append(contactgroup.Numbers, batch...)
					if err := cls.contactGroupDataSource().SaveContactGroup(contactgroup); err != nil {
						return err
					}
					statMap[ListTotal] += int64(len(batch))
					batch = model.NumberList{}
					rnumber = int(cmlconstants.ConfigContactListBatchSize - contactgroup.TotalNumbers)
				}

				if contactgroup.TotalNumbers == cmlconstants.ConfigContactListBatchSize {

					cg, err := cls.ProcessContactListBatch(&ProcessContactListBatchOptions{
						ContactList: op.ContactList,
						UserInfo:    op.UserInfo,
						Batch:       batch,
						Writer:      writer,
					})
					if err != nil {
						logrus.Errorln(fmt.Sprintf("processContactListBatch => contact list Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), err.Error()))
						return err
					}

					if cg != nil {
						statMap[ListTotal] += int64(len(batch))
						cgRef := model.ContactGroupRef{ContactGroupId: cg.ID}
						cgRefs = append(cgRefs, cgRef)
					}
					batch = model.NumberList{}
					rnumber = int(cmlconstants.ConfigContactListBatchSize - contactgroup.TotalNumbers)
				}
			}
		}

	}

	op.ContactList.ContactGroups = cgRefs
	op.ContactList.Imported = statMap[Imported] + op.ContactList.Imported
	op.ContactList.NumberCount = statMap[ListTotal] + op.ContactList.NumberCount
	op.ContactList.Invalid = statMap[Invalid] + op.ContactList.Invalid
	op.ContactList.CellPhone = statMap[CellPhone]
	op.ContactList.Duplicate = statMap[Duplicate]
	op.ContactList.LandLine = statMap[LandLine]
	op.ContactList.Unknown = statMap[Unknown]
	op.ContactList.DNCNumbers = 0

	if op.ContactList.Type == model.ContactListTypeDefault && op.ContactList.ScrubDNC {
		op.ContactList.Status = model.ContactListStatusProcessing
		op.ContactList.DNCScrubStatus = model.DNCScrubProcessStatusAllBatchesSent
		op.ContactList.LandLine = 0
		op.ContactList.CellPhone = 0
	} else {
		op.ContactList.Status = model.ContactListStatusActive
	}

	if err := cls.SaveContactList(op.ContactList); err != nil {
		logrus.Errorln(fmt.Sprintf("saveContactList => contact list Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), err.Error()))
		return err
	}
	// save newly created file to s3
	dstpath, err := cls.GenerateCSV(*op.ContactList)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("generateCSV => contact list Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), err.Error()))
		return err
	}

	s3FilePath, err := cls.uploadContactListFile(op.UserInfo.ID, op.ContactList.ID, dstpath)
	if err == nil {
		op.ContactList.FileName = outFileName
		op.ContactList.FileS3Path = s3FilePath
		// TODO: delete previously uploaded file from s3
	}
	op.ContactList.Status = model.ContactListStatusActive
	if err := cls.SaveContactList(op.ContactList); err != nil {
		logrus.Errorln(fmt.Sprintf("saveContactList => contact list Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), err.Error()))
		return err
	}
	return nil
}

func (cls *ContactListService) uploadContactListFile(userId bson.ObjectId, contId bson.ObjectId, path string) (string, error) {
	destinationPath := "/account/" + userId.Hex() + "/contactlist/" + contId.Hex() + "/"
	s3FilePath, err := cls.S3Service.Upload(destinationPath, path)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("s3fileUpload Error => contact list Id : %s :: Error :%s ::", contId.Hex(), err.Error()))
		return "", err
	}

	return s3FilePath, err
}

func (cls *ContactListService) ProcessContactListBatch(op *ProcessContactListBatchOptions) (*model.ContactGroup, error) {
	// get new instance of contact group datasource
	cgds := cls.contactGroupDataSource()
	defer cgds.Session.Close()

	if op.ContactList.Type == model.ContactListTypeDefault && op.ContactList.ScrubDNC {
		if err := cls.DNCScrubContactList(op); err != nil {
			logrus.Errorln(fmt.Sprintf("Error  DNCScrubContactList  => contact list Id : %s ::user Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), op.UserInfo.ID.Hex(), err.Error()))
			return nil, err
		}
		return nil, nil
	}

	cg := model.ContactGroup{}
	cg.ID = bson.NewObjectId()
	cg.ContactListId = op.ContactList.ID
	cg.Numbers = op.Batch
	err := cgds.SaveContactGroup(&cg)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("saveContactGroup => contact list Id : %s :: contact group Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), cg.ID, err.Error()))
		return nil, err
	}
	defer op.Writer.Flush()
	for _, num := range op.Batch {
		err := op.Writer.Write([]string{num.Number})
		if err != nil {
			logrus.Errorln(fmt.Sprintf("file write failed => contact list Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), err.Error()))

			continue
		}
	}

	return &cg, nil
}

func extractKeys(data map[string]interface{}) (res []string) {
	for key, _ := range data {
		res = append(res, key)
	}

	return
}

func (cls *ContactListService) DNCScrubContactList(op *ProcessContactListBatchOptions) error {

	// get new instance of dnc job datasource
	dncJobds := cls.dncJobDataSource()
	defer dncJobds.Session.Close()

	var chunkSize = cmlconstants.ConfigContactListBatchSize
	var errs []error
	var divided = split(op.Batch, chunkSize)

	m := make(map[string]interface{})
	m["contactListId"] = op.ContactList.ID.Hex()
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

			jobId, err := cls.TCPAService.IsNumberInTCPA(numbersInChunk, m)
			if err != nil {
				logrus.Errorln(fmt.Sprintf("IsNumberInTCPA  => contact list Id : %s ::user Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), op.UserInfo.ID.Hex(), err.Error()))
				errs = append(errs, err)
			}

			job := model.DNCJobs{
				Status:        model.DNCScrubJobTypeProcessing,
				ContactListId: &op.ContactList.ID,
				TTSListID:     nil,
				JobId:         jobId,
			}

			err = dncJobds.SaveDNCJob(&job)
			if err != nil {

				logrus.Errorln(fmt.Sprintf("SaveDNCJob => contact list Id : %s ::user Id : %s  ::job Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), op.UserInfo.ID.Hex(), jobId, err.Error()))
				errs = append(errs, err)
			}
		}()
	}

	wg.Wait()

	if len(errs) > 0 {
		logrus.Errorln(fmt.Sprintf("IsNumberInTCPA  => contact list Id : %s ::user Id : %s :: Error :%s ::", op.ContactList.ID.Hex(), op.UserInfo.ID.Hex(), errs[0].Error()))
		// TODO: compose proper error from all errors and return
		return errs[0]
	}

	return nil
}

// helper
func split(buf []model.Number, lim int) [][]model.Number {
	var chunk []model.Number
	chunks := make([][]model.Number, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:len(buf)])
	}
	return chunks
}

func splittts(buf []model.TTSNumber, lim int) [][]model.TTSNumber {
	var chunk []model.TTSNumber
	chunks := make([][]model.TTSNumber, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:len(buf)])
	}
	return chunks
}

func (cls *ContactListService) CreateNumberInfo(number string) *model.Number {
	numType := cls.ClassifyNumberType(number)
	tz := cls.ClassifyNumberTimeZone(number)
	numTypeStr := ""
	// will eventually deprecate, its here because of design pattern mistake
	switch numType {
	case model.NumberTypeLandLine:
		numTypeStr = "Lindlinenumber"
	case model.NumberTypeCellPhone:
		numTypeStr = "cellnumber"
	}

	return &model.Number{
		ID:            bson.NewObjectId(),
		Number:        number,
		TimeZone:      tz,
		NumberTypeStr: numTypeStr,
		NumberType:    numType,
	}
}

func (cls *ContactListService) WirelessNumber(number string) bool {
	s := cls.RedisClient.HExists("cellnumber", number)
	str := s.String()
	if strings.Contains(str, "true") {
		return true
	} else {
		return false
	}
}

func (cls *ContactListService) LandLineNumber(number string) bool {
	s := cls.RedisClient.HExists("landlinenumber", number)
	str := s.String()
	if strings.Contains(str, "true") {
		return true
	} else {
		return false
	}
}

func (cls *ContactListService) Timezone(number string) string {
	s := cls.RedisClient.HGet("timezone", number)
	str := s.String()
	timezone := strings.Split(str, ":")
	tz := strings.TrimSpace(timezone[1])
	logrus.Infoln(tz)
	return tz
}

// private method
func (cls *ContactListService) extractNumberFromRow(sourceRecords []string, numberColumn int8) (string, error) {
	for i := 0; i < len(sourceRecords); i++ {
		csvRecord := sourceRecords[i]

		if i == int(numberColumn) {
			phoneNumber := cmlutils.SimplifyPhoneNumber(csvRecord)
			return phoneNumber, nil
		}
	}

	return "", errors.New("No Number found")
}

// private method
func (cls *ContactListService) extractTextFromRow(sourceRecords []string, textColumn int8) (string, error) {
	for i := 0; i < len(sourceRecords); i++ {
		csvRecord := sourceRecords[i]

		if i == int(textColumn) {
			return url.QueryEscape(csvRecord), nil
		}
	}

	return "", errors.New("No Text found")
}

func (cls *ContactListService) ClassifyNumberType(number string) model.NumberType {
	if len(number) < 10 {
		return model.NumberTypeUnknown
	}
	num := cmlutils.ExtractNumberFromUSNumber(number)
	prefixSix := num[0:6]

	if cls.WirelessNumber(prefixSix) {
		return model.NumberTypeCellPhone
	} else if cls.LandLineNumber(prefixSix) {
		return model.NumberTypeLandLine
	} else {
		return model.NumberTypeUnknown
	}
}

func (cls *ContactListService) ClassifyNumberTimeZone(number string) string {
	if len(number) < 10 {
		return ""
	}
	num := cmlutils.ExtractNumberFromUSNumber(number)
	prefixThree := num[0:3]
	return cls.Timezone(prefixThree)
}

func (cls *ContactListService) UpdateS3FilePath(contactList *model.ContactList, s3Path string) error {
	// get new instance of sound file datasource
	contactListDataSource := cls.contactListDataSource()
	defer contactListDataSource.Session.Close()

	cl, err := contactListDataSource.ContactListWithId(contactList.ID)
	if err != nil {
		logrus.Errorln("Contact list UpdateS3FilePath failed => Contact list Id : " + contactList.ID.Hex())
		return err
	}

	// update s3 Path
	cl.FileS3Path = s3Path

	return contactListDataSource.SaveContactList(cl)
}

func (cls *ContactListService) DeleteContactList(contactList *model.ContactList) error {
	// get new instance of contact list datasource
	contactListDataSource := cls.contactListDataSource()
	defer contactListDataSource.Session.Close()

	// get new instance of contact group datasource
	contactGroupDataSource := cls.contactGroupDataSource()
	defer contactGroupDataSource.Session.Close()

	// Delete all associated contact groups as well
	if err := contactGroupDataSource.DeleteContactGroupWithContactListId(contactList.ID); err != nil {
		logrus.Errorln("Contact Groups delete failed => Contact list Id : " + contactList.ID.Hex())
		return err
	}

	return contactListDataSource.DeleteContactList(contactList)
}

func (cls *ContactListService) IsValidContactListFileFormat(fileName string) bool {
	isValidFileFormat := false

	if strings.HasSuffix(fileName, ".csv") {
		isValidFileFormat = true
	}

	if strings.HasSuffix(fileName, ".xls") {
		isValidFileFormat = true
	}

	if strings.HasSuffix(fileName, ".xlsx") {
		isValidFileFormat = true
	}

	return isValidFileFormat
}

func (cls *ContactListService) IsContactListXlx(fileName string) bool {
	isValidFileFormat := false

	if strings.HasSuffix(fileName, ".xls") {
		isValidFileFormat = true
	}

	if strings.HasSuffix(fileName, ".xlsx") {
		isValidFileFormat = true
	}

	return isValidFileFormat
}

func (cls *ContactListService) GetColumnIndexFromNumberColumnValue(numberColumnValue string) int8 {
	reference := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

	for i := 0; i < len(reference); i++ {
		item := reference[i]
		if item == numberColumnValue {
			return int8(i)
		}
	}

	return -1
}

func (cls *ContactListService) NumbersToCSV(numbers []string, name string) (*os.File, error) {

	file, err := os.Create(name + ".csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, v := range numbers {

		err := writer.Write(strings.Fields(v))
		if err != nil {
			logrus.Errorln(fmt.Sprintf("file write failed => contact list name : %s :: Error :%s ::", name, err.Error()))
			return nil, err
		}
	}

	return file, nil
}

// this method saves given to a destination path
// it returns destination file path in case of success
// in case of failure it returns error

func (cls *ContactListService) ReadFile(file string) (*os.File, *csv.Reader, error) {
	filein, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}
	inFileReader := csv.NewReader(filein)
	return filein, inFileReader, nil

}

func (cls *ContactListService) SaveTTSFile(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", errors.New(cmlmessages.ContactListReadError)
	}
	defer src.Close()

	basePath := filepath.Base(file.Filename)
	ext := filepath.Ext(file.Filename)

	fn, err := cmlutils.RemoveNonAlphaNumeric(basePath)
	if err != nil {
		return "", errors.New(cmlmessages.ContactListWriteError)
	}

	fileName := fn + ext

	// Destination
	destinationPath := cmlconstants.TempDestinationPath + fileName
	// prepare the dst
	os.Mkdir(cmlconstants.TempDestinationPath, os.ModePerm)

	dst, err := os.Create(destinationPath)
	if err != nil {
		return "", errors.New(cmlmessages.ContactListWriteError)
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return "", errors.New(cmlmessages.ContactListWriteError)
	}

	return destinationPath, nil

}

func (cls *ContactListService) SaveMultipartFile(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", errors.New(cmlmessages.ContactListReadError)
	}
	defer src.Close()

	basePath := filepath.Base(file.Filename)
	ext := filepath.Ext(file.Filename)

	fn, err := cmlutils.RemoveNonAlphaNumeric(basePath)
	if err != nil {
		return "", errors.New(cmlmessages.ContactListWriteError)
	}

	fileName := fn + ext

	// Destination
	destinationPath := cmlconstants.TempDestinationPath + fileName

	// prepare the dst
	os.Mkdir(cmlconstants.TempDestinationPath, os.ModePerm)

	dst, err := os.Create(destinationPath)
	if err != nil {
		return "", errors.New(cmlmessages.ContactListWriteError)
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return "", errors.New(cmlmessages.ContactListWriteError)
	}

	if cls.IsContactListXlx(fileName) {
		logrus.Infoln("file format is xlsx")

		fileBaseName := filepath.Base(destinationPath)
		fileName := strings.TrimSuffix(fileBaseName, filepath.Ext(fileBaseName))

		newDestinationPath := cmlconstants.TempDestinationPath + fileName + ".csv"
		// convert xls to csv

		command := fmt.Sprintf("ssconvert %s %s", destinationPath, newDestinationPath)
		logrus.Infoln("new destination path : " + newDestinationPath + "\n")
		logrus.Infoln("command : " + command + "\n")

		result, error2 := exec.Command("sh", "-c", command).Output()
		if error2 != nil {
			cmlutils.DeleteFile(destinationPath)
			logrus.Errorln(error2.Error())
			logrus.Infoln("\n")
			return "", errors.New(cmlmessages.ContactListWriteError)
		}

		cmlutils.DeleteFile(destinationPath)

		fmt.Printf("%s\n\n", result)
		return newDestinationPath, nil
	}

	return destinationPath, nil
}

func (cls *ContactListService) HasNumberDNC(userId string, number string, limit int, page int) ([]model.DncNumber, int) {
	var list []model.DncNumber
	var CGroupDeatail model.ContactGroups
	if !bson.IsObjectIdHex(userId) {
		return nil, 0
	}

	cl, err := cls.DoNotContactListWithUserId(userId)
	if err != nil {
		return nil, 0
	}
	// get new instance of contact list datasource
	contactListDataSource := cls.contactListDataSource()
	defer contactListDataSource.Session.Close()

	// get new instance of contact group datasource
	contactGroupDataSource := cls.contactGroupDataSource()
	defer contactGroupDataSource.Session.Close()
	logrus.Infoln("contactlist", cl)

	//hasNumber := false
	var count int
	var err1 error
	// iterate over all contact groups
	for _, cont := range cl {
		for _, cgr := range cont.ContactGroups {

			CGroupDeatail, _, err1 = contactGroupDataSource.HasNumber(cgr.ContactGroupId, number, limit, page)
			if err1 != nil {
				return nil, 0
			}
			if len(CGroupDeatail) > 0 {
				logrus.Infoln(CGroupDeatail)
				count++
				data := model.DncNumber{cont.Name, cont.ID, cgr.ContactGroupId, number}
				list = append(list, data)
			} else {
				continue
			}

		}

	}
	return list, count
}

func (cls *ContactListService) RemoveNumberFromDNC(userID string, number string) error {
	if !bson.IsObjectIdHex(userID) {
		return errors.New("object id is not valid")
	}

	//	cl, err := cls.DoNotContactListWithUserId(userID)
	//	if err != nil {
	//		return err
	//	}

	// get new instance of contact list datasource
	contactListDataSource := cls.contactListDataSource()
	defer contactListDataSource.Session.Close()

	// get new instance of contact group datasource
	contactGroupDataSource := cls.contactGroupDataSource()
	defer contactGroupDataSource.Session.Close()

	//	totalCount := int64(0)
	//	success := true
	//	totalUpdated := 0
	// iterate over all contact groups
	/*	for _, cgr := range cl.ContactGroups {

		// remove number from contact group
		updated, err := contactGroupDataSource.PullNumberFromContactGroup(cgr.ContactGroupId, number)
		if err != nil {
			return err
		}

		// update count of total numbers removed
		totalUpdated = totalUpdated + updated

		// get updated contact group
		cgNew, err := contactGroupDataSource.ContactGroupWithId(cgr.ContactGroupId)
		if err != nil {
			success = false
			continue
		}*/

	// update numbers count for contact group
	/*	if err := contactGroupDataSource.UpdateNumberCount(cgNew.ID, int32(len(cgNew.Numbers))); err != nil {
				success = false
				continue
			}

			totalCount = totalCount + int64(len(cgNew.Numbers))
		}

		// save total number count for dnc
		if success && totalUpdated > 0 {
			return contactListDataSource.UpdateDNCTotalNumberCount(bson.ObjectIdHex(userID), totalCount)
		} else {
			contactListDataSource.UpdateDNCTotalNumberCount(bson.ObjectIdHex(userID), totalCount)
			return errors.New("no dnc number found")
		}*/
	return nil
}

func (cls *ContactListService) GenerateCSV(contactList model.ContactList) (string, error) {

	// get new instance of contact group datasource
	contactGroupDataSource := cls.contactGroupDataSource()
	defer contactGroupDataSource.Session.Close()

	dstPath := "/tmp/" + contactList.FileName + ".csv"

	// create csv file to put data to
	file, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}

	defer file.Close()

	// generate writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, cgRef := range contactList.ContactGroups {
		cg, err := contactGroupDataSource.ContactGroupWithId(cgRef.ContactGroupId)
		if err != nil {
			return "", err
		}

		for _, number := range cg.Numbers {
			// generate csv
			err := writer.Write([]string{number.Number})
			if err != nil {
				return "", err
			}
		}
	}

	return dstPath, nil
}

// return instance if contact list  data source
// every time a new instance would be created
func (cls *ContactListService) contactListDataSource() *datasource.ContactListDataSource {
	return &datasource.ContactListDataSource{DataSource: datasource.DataSource{Session: cls.Session.Copy()}}
}

// return instance if dnc job  data source
// every time a new instance would be created
func (cls *ContactListService) dncJobDataSource() *datasource.DNCJobsDataSource {
	return &datasource.DNCJobsDataSource{DataSource: datasource.DataSource{Session: cls.Session.Copy()}}
}

// return instance if dnc job result data source
// every time a new instance would be created
func (cls *ContactListService) dncJobResultDataSource() *datasource.DNCJobsResultDataSource {
	return &datasource.DNCJobsResultDataSource{DataSource: datasource.DataSource{Session: cls.Session.Copy()}}
}

// return instance if contact group  data source
// every time a new instance would be created
func (cls *ContactListService) contactGroupDataSource() *datasource.ContactGroupDataSource {
	return &datasource.ContactGroupDataSource{DataSource: datasource.DataSource{Session: cls.Session.Copy()}}
}

func (cgs *ContactListService) IncrementContactListNumberCount(objectID bson.ObjectId) error {

	// get new instance of sound file datasource
	contactListDataSource := cgs.contactListDataSource()
	defer contactListDataSource.Session.Close()

	return contactListDataSource.IncrementContactListNumberCount(objectID)
}

func (cgs *ContactListService) DecrementContactListNumberCount(objectID bson.ObjectId) error {

	// get new instance of sound file datasource
	contactListDataSource := cgs.contactListDataSource()
	defer contactListDataSource.Session.Close()

	return contactListDataSource.DecrementContactListNumberCount(objectID)
}

// private method
func createContactGroupObject(contactList *model.ContactList) *model.ContactGroup {
	contactGroup := model.ContactGroup{}

	contactGroup.ID = bson.NewObjectId()
	contactGroup.ContactListId = contactList.ID
	contactGroup.Numbers = model.NumberList{}

	return &contactGroup
}
