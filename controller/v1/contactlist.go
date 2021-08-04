package controllerv1

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlconstants"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/network/response"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlmessages"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

type ContactListController struct {
	//Controller
	ContactListService         *service.ContactListService
	ContactGroupService        *service.ContactGroupService
	UserService                *service.UserService
	UserDetailService          *service.UserDetailService
	S3Service                  *service.S3Service
	TCPAService                *service.TCPAService
	CampaignService            *service.CampaignService
	NumbersCarrierGroupService *service.NumbersCarrierGroupService
	DNCJobsService             *service.DNCJobService
	FileService                *service.FileService
	DNCJobResultService        *service.DNCJobResultService
}

type (
	AddContactNumberRequest struct {
		Number string `json:"number" validate:"required"`
	}
	ContactListIDs struct {
		ID []string `json:"id"`
	}
)

type CallerGroupPurchasedNumberRequest struct {
	Numbers []string `json:"numbers"`
	Name    string   `json:"name"`
}

func (contCont *ContactListController) CallerGroupFromPurchasedNumbers(c echo.Context) error {
	cc := c.(model.AppContext)

	var req CallerGroupPurchasedNumberRequest

	if err := c.Bind(&req); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Please provide all required fields")
	}

	userInfo, errUser := contCont.UserService.UserObject(cc.ID.Hex())
	if errUser != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	file, err := contCont.ContactListService.NumbersToCSV(req.Numbers, req.Name)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	cl, err := contCont.ContactListService.SaveContactListNew(&service.ContactListSaveOptions{
		Name:            req.Name,
		FileName:        file.Name(),
		FilePath:        filepath.Base(file.Name()),
		UserInfo:        userInfo,
		NumberColumn:    0,
		HeaderPresent:   false,
		ContactListType: model.ContactListTypeCallerGroups,
		RemoveDuplicate: false,
		ShuffleNumber:   false,
		ScrubLandline:   false,
		ScrubCellPhone:  false,
		AreaCodeDialing: true,
		ScrubDNC:        false,
		Randomize:       false,
	})

	//Marshal provided interface into JSON structure
	uj, _ := json.Marshal(cl)

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "contact list uploaded success", "contact list uploaded successfully", uj)
}

func (contCont *ContactListController) UploadContactListForUser(c echo.Context) error {
	cc := c.(model.AppContext)

	// form values
	numColfv := c.FormValue("numberColumn")
	isHeaderPresentfv := c.FormValue("isHeaderPresent")
	remDupfv := c.FormValue("removeDup")
	scrubDNCfv := c.FormValue("scrubDNC")
	shufflefv := c.FormValue("shuffle")
	scrubLndLivefv := c.FormValue("scrubLandLine")
	scrubCellPhonefv := c.FormValue("scrubCellPhone")
	areaCodeDialingfv := c.FormValue("areaCode")
	randomizefv := c.FormValue("random")
	cltfv := c.FormValue("type")
	namefv := c.FormValue("name")

	// process form values
	headerPresent := false
	var textCol int8 = -1
	headerPresent, _ = strconv.ParseBool(isHeaderPresentfv)
	clt, err := strconv.Atoi(cltfv)
	if err != nil {
		clt = int(model.ContactListTypeDefault)
	}

	// extract value of number column
	numCol := contCont.ContactListService.GetColumnIndexFromNumberColumnValue(numColfv)
	if numCol < 0 {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListNumberColumnIncorrect)
	}

	userInfo, errUser := contCont.UserService.UserObject(cc.ID.Hex())
	if errUser != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListSelectFile)
	}
	// check file format
	if !contCont.ContactListService.IsValidContactListFileFormat(file.Filename) {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListFormatInCorrect)
	}

	// save incoming file
	// TODO: look where in code this file would be deleted from server
	filePath, errSave := contCont.ContactListService.SaveMultipartFile(file)
	if errSave != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, errSave.Error())
	}

	filein, reader, err := contCont.ContactListService.ReadFile(filePath)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, errSave.Error())
	}

	record, err := reader.Read()
	if err != nil {
		cmlutils.DeleteFile(filePath)
		filein.Close()
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, err.Error())
	}
	if len(record) <= int(numCol) || len(record) <= int(textCol) {
		cmlutils.DeleteFile(filePath)
		filein.Close()
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, "Please Provide Valid Column Number or Text Number")
	}
	filein.Close()
	remDup, _ := strconv.ParseBool(remDupfv)
	shuffleNum, _ := strconv.ParseBool(shufflefv)
	scrubLndLine, _ := strconv.ParseBool(scrubLndLivefv)
	scrubCellPhone, _ := strconv.ParseBool(scrubCellPhonefv)
	areaCodeDialing, _ := strconv.ParseBool(areaCodeDialingfv)
	scrubDNC, _ := strconv.ParseBool(scrubDNCfv)
	randomize, _ := strconv.ParseBool(randomizefv)

	cl, err := contCont.ContactListService.SaveContactListNew(&service.ContactListSaveOptions{
		Name:            namefv,
		FileName:        file.Filename,
		FilePath:        filePath,
		UserInfo:        userInfo,
		NumberColumn:    numCol,
		TextColumn:      textCol,
		HeaderPresent:   headerPresent,
		ContactListType: model.ContactListType(clt),
		RemoveDuplicate: remDup,
		ShuffleNumber:   shuffleNum,
		ScrubLandline:   scrubLndLine,
		ScrubCellPhone:  scrubCellPhone,
		AreaCodeDialing: areaCodeDialing,
		ScrubDNC:        scrubDNC,
		Randomize:       randomize,
	})

	//Marshal provided interface into JSON structure
	uj, _ := json.Marshal(cl)

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "contact list uploaded success", "contact list uploaded successfully", uj)
}

func (contCont *ContactListController) processUpdateContactList(contactId string, s3filepath string, op *service.ContactListSaveOptions) (*model.ContactList, int, error) {
	contactList := model.ContactList{}
	fileservice := service.FileService{}
	//contactList.ID = bson.NewObjectId()
	var invalidnumber int
	var numberslice []string
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

	var dupMap = make(map[string]bool)

	contactList.UserID = op.UserInfo.ID
	contactList.ID = bson.ObjectIdHex(contactId)
	list, err := contCont.ContactListService.ContactListObjectWithIdAndUserId(contactList.ID.Hex(), op.UserInfo.ID)
	if err != nil {
		return nil, 0, err
	}
	contactList.FileName = list.FileName
	contactList.Name = list.Name
	contactList.Type = list.Type
	contactList.FileS3Path = list.FileS3Path
	contactList.CreateDate = list.CreateDate
	if contactList.Type == model.ContactListTypeDNC {
		contactList.AreaCode = false
		contactList.ScrubLandLine = false
		contactList.Shuffle = false
		contactList.ScrubCellPhone = false
		contactList.RemoveDup = true

	}
	var lastId model.ContactGroupRef
	if len(list.ContactGroups) > 0 {
		lastId = list.ContactGroups[len(list.ContactGroups)-1]
	} else {
		createContactGroupObject(&contactList)
	}
	contactgroup, err := contCont.ContactGroupService.ContactGroupWithId(lastId.ContactGroupId.Hex())
	if err != nil {
		contactgroup = createContactGroupObject(&contactList)
		//return nil, err
	}
	file, err := os.Open(op.FilePath)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	r := csv.NewReader(file)
	var numberList = model.NumberList{}
	r.FieldsPerRecord = -1
	recordCount := 0
	batchCount := 0

	listCount := 0

	// declare variable to keep all the created contact group references
	//	contactGroupsReferences := model.ContactGroupRefs{}
	if rec, err := r.Read(); err != nil { //read header
		return nil, 0, errors.New(cmlmessages.ContactListReadError)
	} else {
		if !(int8(len(rec)) > op.NumberColumn) {
			// column number value exceeds that of actual column
			return nil, 0, errors.New("number column value exceeds that of actual columns present in file")
		}

		if op.HeaderPresent == true {
			// there is header present in first row
			// skip that
		} else {
			number, err := extractNumber(rec, int(op.NumberColumn))
			if err == nil {
				numberslice = append(numberslice, number.Number)
				if cmlutils.IsUSNumber(number.Number) == true {
					// here extract last 10 digits from the number

					numInfo := contCont.ContactListService.CreateNumberInfo(number.Number)

					if numInfo.NumberType == model.NumberTypeCellPhone {
						recordCount++
						statMap[CellPhone] += 1
					} else if numInfo.NumberType == model.NumberTypeLandLine {
						recordCount++
						statMap[LandLine] += 1
					} else {
						recordCount++
						statMap[Unknown] += 1
					}

					if len(numberList) == 0 {
						if contactList.ScrubCellPhone == true && numInfo.NumberType == model.NumberTypeLandLine {

						} else if contactList.ScrubLandLine == true && numInfo.NumberType == model.NumberTypeCellPhone {

						} else {

							numberList = append(numberList, *numInfo)
							listCount++
							batchCount++
						}
					} else {

						if contactList.ScrubCellPhone == true && numInfo.NumberType == model.NumberTypeCellPhone {

						} else if contactList.ScrubLandLine == true && numInfo.NumberType == model.NumberTypeLandLine {

						} else {
							numberList = append(numberList, *numInfo)
							listCount++
							batchCount++
						}
					}
				} else {
					statMap[Invalid] += 1
				}
			}
		}
	}
	for {
		rec, err := r.Read()
		if err != nil {
			if err == io.EOF {
				logrus.Infoln("EOF reached")

				break
			} else {
				logrus.Infoln("Error occured")
			}

		} else {
			number, err := extractNumber(rec, int(op.NumberColumn))
			if err == nil {
				numberslice = append(numberslice, number.Number)

				if cmlutils.IsUSNumber(number.Number) == true {
					// here extract last 10 digits from the number
					numInfo := contCont.ContactListService.CreateNumberInfo(number.Number)
					if numInfo.NumberType == model.NumberTypeCellPhone {
						recordCount++
						statMap[CellPhone] += 1
					} else if numInfo.NumberType == model.NumberTypeLandLine {
						recordCount++
						statMap[LandLine] += 1
					} else {
						recordCount++
						statMap[Unknown] += 1
					}
					_, ok := dupMap[numInfo.Number]
					if ok {
						// duplicate
						statMap[Duplicate] += 1
						continue
					} else {
						dupMap[numInfo.Number] = true
					}
					if contactList.ScrubLandLine == true && numInfo.NumberType == model.NumberTypeLandLine {

						continue
					}
					if contactList.ScrubCellPhone == true && numInfo.NumberType == model.NumberTypeCellPhone {

						continue
					}

					if len(numberList) == 0 {
						numberList = append(numberList, *numInfo)
						listCount++
						batchCount++
					} else {
						numberList = append(numberList, *numInfo)
						listCount++
						batchCount++
					}

				} else {
					invalidnumber++
				}
			}
			var rnumber = int(cmlconstants.ConfigContactListBatchSize - contactgroup.TotalNumbers)
			if contactgroup.TotalNumbers < cmlconstants.ConfigContactListBatchSize {
				if batchCount == rnumber {
					contactgroup.Numbers = append(contactgroup.Numbers, numberList...)
					if err := contCont.ContactGroupService.SaveContactGroup(contactgroup); err != nil {
						return nil, 0, err
					}
				}

			}
			if contactgroup.TotalNumbers == cmlconstants.ConfigContactListBatchSize {
				contactgroup = createContactGroupObject(&contactList)
				contactgroup.Numbers = append(contactgroup.Numbers, *number)

				if err := contCont.ContactGroupService.SaveContactGroup(contactgroup); err != nil {
					return nil, 0, err
				}

				numberList = model.NumberList{}
				rnumber = int(cmlconstants.ConfigContactListBatchSize - contactgroup.TotalNumbers)
				batchCount = 0
				// cache contact group info to attach in contact list later
				contactGroupRef := model.ContactGroupRef{ContactGroupId: contactgroup.ID}
				list.ContactGroups = append(list.ContactGroups, contactGroupRef)
			}
		}
	}

	if contactList.Shuffle == true && contactList.Type == model.ContactListTypeDefault {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(numberList), func(i, j int) { numberList[i], numberList[j] = numberList[j], numberList[i] })
	}
	if len(numberList) > 0 {
		//	contactGroup := createContactGroupObject(&contactList)
		contactgroup.Numbers = append(contactgroup.Numbers, numberList...)
		//		contactgroup.TotalNumbers = int32(len(numberList))

		if err := contCont.ContactGroupService.SaveContactGroup(contactgroup); err != nil {
			return nil, 0, err
		}
	}

	contactList.ContactGroups = list.ContactGroups
	list.NumberCount += int64(listCount)
	contactList.Invalid = statMap[Invalid]
	contactList.Duplicate = statMap[Duplicate]
	if contactList.ScrubLandLine != true {
		contactList.LandLine = statMap[LandLine]
	}
	if contactList.ScrubCellPhone != true {
		contactList.CellPhone = statMap[CellPhone]
	}

	if contactList.Type == model.ContactListTypeDefault && contactList.RemoveDup == true {
		contactList.Imported = int64(recordCount) + contactList.Duplicate + contactList.Invalid
	} else {
		contactList.Imported = int64(recordCount) + contactList.Invalid
	}

	contactList.NumberCount = list.NumberCount

	fileid, _ := cmlutils.GetUUID()
	filepath, err4 := fileservice.MerageFile(fileid+"_"+contactList.ID.Hex(), op.FilePath, s3filepath)
	if err4 != nil {
		return nil, 0, err4
	}
	s3FilePath, err := contCont.uploadContactListFile(contactList.UserID, contactList.ID, filepath)
	if err == nil {
		contactList.FileName = contactList.Name
		contactList.FileS3Path = s3FilePath
	}

	contactList.Status = model.ContactListStatusActive
	if err := contCont.ContactListService.SaveContactList(&contactList); err != nil {
		logrus.Errorln("Error Save Contact list")
		logrus.Errorln(contactList)
		return nil, 0, err
	}
	defer cmlutils.DeleteFile(filepath)
	return &contactList, 0, nil
}

func (contCont *ContactListController) UpdateContactNameForContactList(c echo.Context) error {
	userID := c.Param("id")
	contactListId := c.Param("cid")

	// Stub an user to be populated from the body
	contactListNameObject := NameUpdateObject{}

	if err := c.Bind(&contactListNameObject); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Please provide all required fields")
	}

	// validate input request body
	if err := c.Validate(contactListNameObject); err != nil {
		logrus.Debugln(err)
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	// validate if contact name is empty
	if contactListNameObject.Name == "" {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListProvideAllFields)
	}

	// get user object
	user, er1 := contCont.UserService.UserObject(userID)
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	// get contact list
	contactListObject, er1 := contCont.ContactListService.ContactListObjectWithId(contactListId)
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListDoesNotExist)
	}

	if user.ID.Hex() != contactListObject.UserID.Hex() {
		// given user is not authorized to make this call
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, "Sorry! You don't have access rights to make this change", "")
	}

	// update contact list name
	contactListObject.Name = contactListNameObject.Name

	// save the updated change
	if err := contCont.ContactListService.SaveContactList(contactListObject); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	// Marshal provided interface into JSON structure
	sf, _ := json.Marshal(contactListObject)

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Name update", "Name updated", sf)
}

// get list of all contact lists
// this method also supports server side pagination
func (contCont *ContactListController) GetContactListsForUser(c echo.Context) error {
	cc := c.(model.AppContext)
	// Grab user id for which file is being uploaded
	ctype := c.QueryParam("type") // convert to int and send below
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	fetchAll := c.QueryParam("listall")

	fetchAllBool, err := strconv.ParseBool(fetchAll)
	if err != nil {
		fetchAllBool = false
	}

	pageNumber, errPage := strconv.Atoi(page)
	limitNumber, errLimit := strconv.Atoi(limit)

	if errPage != nil || errLimit != nil {
		pageNumber = 1
		limitNumber = 0
	}
	logrus.Debugln(ctype)
	var contacttype int8
	contype, err := strconv.Atoi(ctype)
	logrus.Debugln(contype)
	contacttype = int8(contype)
	logrus.Debugln(contacttype)
	list, count, err := contCont.ContactListService.ContactListListingWithUserId(cc.ID.Hex(), pageNumber, limitNumber, contacttype, fetchAllBool) // remove isDonotContactListValue
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	pageInfo := model.Page{}
	pageInfo.Page = pageNumber
	pageInfo.Limit = limitNumber
	pageInfo.TotalCount = count

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(list)
	pageData, _ := json.Marshal(pageInfo)

	return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusOK, "Contact List fetched successfully", "Contact List fetched successfully", data, pageData)
}

func (contCont *ContactListController) ContactListDeleteMultiple(c echo.Context) error {
	cc := c.(model.AppContext)
	// Grab user id for which file is being uploaded
	//contactListInfoId := c.Param("cid")
	contactlistObjects := ContactListIDs{}
	var count int

	if err := c.Bind(&contactlistObjects); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	user, er1 := contCont.UserService.UserObject(cc.ID.Hex())
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	for _, i := range contactlistObjects.ID {
		contactListObject, er2 := contCont.ContactListService.ContactListObjectWithId(i)
		if er2 != nil {
			logrus.Errorln("ContactList Object Not Found")
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListDoesNotExist)
		}

		if user.ID.Hex() != contactListObject.UserID.Hex() {
			// given user is not authorized to make this call
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListUnauthorizedForAction)
		}

		bool, err := contCont.CampaignService.IsContactListAttached(cc.ID.Hex(), i)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
		}

		if bool {
			// contact list is attached to some campaign, you can not delete this contact list
			continue
			//return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListAttachedWithCamapign)
		}

		if err := contCont.S3Service.CheckIfObjectExists(cmlutils.S3BucketName(), contactListObject.FileS3Path); err != nil {
			// contact list doesn't exist on server
			if err := contCont.ContactListService.DeleteContactList(contactListObject); err != nil {
				logrus.Errorln("Contact list delete failed => Contact list Id : " + contactListObject.ID.Hex())
				return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListDeleteOperationFailed)
			}
			count++
			// update user object in db
			if err := contCont.UserService.SaveUser(user); err != nil {
				logrus.Errorln("User save failed while deleting contact list => User id : " + user.ID.Hex() + " => Contact List : " + contactListObject.ID.Hex())
			}
			continue
			//continue
			//return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Contact list deleted successfully", "Contact list Deleted successfully")
			//return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListDeleteFailForS3)
		}

		if err := contCont.S3Service.DeleteS3Object(cmlutils.S3BucketName(), contactListObject.FileS3Path); err != nil {
			// delete file from s3
			if err := contCont.ContactListService.DeleteContactList(contactListObject); err != nil {
				logrus.Errorln("Contact list delete failed => Contact list Id : " + contactListObject.ID.Hex())
				return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListDeleteOperationFailed)
			}
			count++
			// update user object in db
			if err := contCont.UserService.SaveUser(user); err != nil {
				logrus.Errorln("User save failed while deleting contact list => User id : " + user.ID.Hex() + " => Contact List : " + contactListObject.ID.Hex())
			}
			continue
			//return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Contact list deleted successfully", "Contact list Deleted successfully")
			//	return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListDeleteFailForS3)
		}

		if err := contCont.ContactListService.DeleteContactList(contactListObject); err != nil {
			logrus.Errorln("Contact list delete failed => Contact list Id : " + contactListObject.ID.Hex())
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListDeleteOperationFailed)
		}
		count++
		// update user object in db
		if err := contCont.UserService.SaveUser(user); err != nil {
			logrus.Errorln("User save failed while deleting contact list => User id : " + user.ID.Hex() + " => Contact List : " + contactListObject.ID.Hex())
		}
	}
	submessage := fmt.Sprintf("%d Contact list Deleted successfully", count)
	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Contact list deleted successfully", submessage)
}

// contact list files can be deleted using this method
func (contCont *ContactListController) ContactListDelete(c echo.Context) error {
	cc := c.(model.AppContext)
	// Grab user id for which file is being uploaded
	contactListInfoId := c.Param("cid")

	user, er1 := contCont.UserService.UserObject(cc.ID.Hex())
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	contactListObject, er2 := contCont.ContactListService.ContactListObjectWithId(contactListInfoId)
	if er2 != nil {
		logrus.Debugln("ContactList Object Not Found")
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListDoesNotExist)
	}

	if user.ID.Hex() != contactListObject.UserID.Hex() {
		// given user is not authorized to make this call
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListUnauthorizedForAction)
	}

	bool, err := contCont.CampaignService.IsContactListAttached(cc.ID.Hex(), contactListInfoId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	if bool {
		// contact list is attached to some campaign, you can not delete this contact list
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListAttachedWithCamapign)
	}

	if err := contCont.S3Service.CheckIfObjectExists(cmlutils.S3BucketName(), contactListObject.FileS3Path); err != nil {
		// contact list doesn't exist on server
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListDeleteFailForS3)
	}

	if err := contCont.S3Service.DeleteS3Object(cmlutils.S3BucketName(), contactListObject.FileS3Path); err != nil {
		// delete file from s3
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListDeleteFailForS3)
	}

	if err := contCont.ContactListService.DeleteContactList(contactListObject); err != nil {
		logrus.Debugln("Contact list delete failed => Contact list Id : " + contactListObject.ID.Hex())
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListDeleteOperationFailed)
	}

	// Check Number Group
	if err := contCont.NumbersCarrierGroupService.DeleteNumberGroupsByContactListID(cc.ID.Hex(), contactListInfoId); err != nil {
		logrus.Debugln("User save failed while deleting contact list => User id : " + user.ID.Hex() + " => Contact List : " + contactListObject.ID.Hex())
	}
	// update user object in db
	if err := contCont.UserService.SaveUser(user); err != nil {
		logrus.Debugln("User save failed while deleting contact list => User id : " + user.ID.Hex() + " => Contact List : " + contactListObject.ID.Hex())
	}

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Contact list deleted successfully", "Contact list Deleted successfully")
}

// contact list files can be deleted using this method
func (contCont *ContactListController) ContactListDetail(c echo.Context) error {

	// Grab user id for which file is being uploaded
	contactListInfoId := c.Param("cid")

	contactListObject, er2 := contCont.ContactListService.ContactListObjectWithId(contactListInfoId)
	if er2 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListDoesNotExist)
	}

	//Marshal provided interface into JSON structure
	uj, _ := json.Marshal(contactListObject)

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "contact list found", "contact list found", uj)
}

// TODO: replace with new implementation
func (contCont *ContactListController) AddDncNumber(c echo.Context) error {
	//// Grab user id for which file is being uploaded
	contactListInfoId := c.Param("cid")
	cc := c.(model.AppContext)
	//
	//// Grab user id for which file is being uploaded
	// Grab user id for which file is being uploaded
	columnNumberValue := c.FormValue("numberColumn")
	isHeaderPresent := c.FormValue("isHeaderPresent")
	remDupfv := c.FormValue("remDup")
	scrubDNCfv := c.FormValue("scrubDNC")
	shufflefv := c.FormValue("shuffle")
	scrubLndLivefv := c.FormValue("scrubLandLine")
	scrubCellPhonefv := c.FormValue("scrubCellPhone")
	areaCodeDialingfv := c.FormValue("areaCode")
	randomizefv := c.FormValue("random")
	ctype := c.FormValue("type")
	name := c.FormValue("name")

	// check value if header is present

	headerPresent := false
	headerPresent, _ = strconv.ParseBool(isHeaderPresent)
	// extract value of number column
	columnNumberValueInt := contCont.ContactListService.GetColumnIndexFromNumberColumnValue(columnNumberValue)
	if columnNumberValueInt < 0 {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListNumberColumnIncorrect)
	}

	userInfo, errUser := contCont.UserService.UserObject(cc.ID.Hex())
	if errUser != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListSelectFile)
	}
	// check file format
	if !contCont.ContactListService.IsValidContactListFileFormat(file.Filename) {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListFormatInCorrect)
	}

	// save incoming file
	destinationPath, errSave := contCont.ContactListService.SaveMultipartFile(file)
	if errSave != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, errSave.Error())
	}

	contype, err := strconv.Atoi(ctype)

	remDup, _ := strconv.ParseBool(remDupfv)
	shuffleNum, _ := strconv.ParseBool(shufflefv)
	scrubLndLine, _ := strconv.ParseBool(scrubLndLivefv)
	scrubCellPhone, _ := strconv.ParseBool(scrubCellPhonefv)
	areaCodeDialing, _ := strconv.ParseBool(areaCodeDialingfv)
	scrubDNC, _ := strconv.ParseBool(scrubDNCfv)
	randomize, _ := strconv.ParseBool(randomizefv)

	contactListObject, er2 := contCont.ContactListService.ContactListObjectWithId(contactListInfoId)
	if er2 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListDoesNotExist)
	}

	fileService := contCont.FileService
	fileid, _ := cmlutils.GetUUID()
	filepath := "s3" + fileid + contactListObject.ID.Hex()
	err2 := fileService.DownloadFile(filepath, os.Getenv("BUCKET_PATH")+contactListObject.FileS3Path)
	if err2 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err2.Error())
	}

	contactListInfo, err1 := contCont.ContactListService.UpdateContactList(contactListInfoId,
		filepath,
		&service.ContactListSaveOptions{
			Name:            name,
			FilePath:        destinationPath,
			FileName:        file.Filename,
			UserInfo:        userInfo,
			ContactListType: model.ContactListType(contype),
			NumberColumn:    columnNumberValueInt,
			HeaderPresent:   headerPresent,
			RemoveDuplicate: remDup,
			ShuffleNumber:   shuffleNum,
			ScrubLandline:   scrubLndLine,
			ScrubCellPhone:  scrubCellPhone,
			AreaCodeDialing: areaCodeDialing,
			ScrubDNC:        scrubDNC,
			Randomize:       randomize,
		})
	if err1 != nil {
		//delete locally saved file
		cmlutils.DeleteFile(destinationPath)
		cmlutils.DeleteFile(filepath)
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err1.Error())
	}
	//	logrus.Infoln(invalidnumber)
	campaigninfo, err := contCont.CampaignService.CampaignObjectWithContactId(contactListInfo.ID.Hex())
	if err != nil {
		logrus.Errorln(err.Error())
	}
	for _, c := range campaigninfo {
		if c.Status == cmlconstants.CampaignStatusRunning || c.Status == cmlconstants.CampaignStatusStopped {

			go func() {
				url := fmt.Sprintf(cmlutils.VOIPAPIPath()+"campaign/%s/dnc", c.ID.Hex())
				resp, err := http.Get(url)
				if err != nil {
					logrus.Errorln(err.Error())
					//return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
				}
				logrus.Infoln(resp.Body)

			}()
		} else {
			continue
		}
	}
	//delete locally saved file
	cmlutils.DeleteFile(destinationPath)
	cmlutils.DeleteFile(filepath)
	//Marshal provided interface into JSON structure
	uj, _ := json.Marshal(contactListInfo)

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "contact list found", "contact list found", uj)

}

// download contact list
func (contCont *ContactListController) DownloadContactList(c echo.Context) error {
	// Grab user id for which file is being uploaded
	userID := c.Param("id")
	contactListInfoId := c.Param("cid")

	user, er1 := contCont.UserService.UserObject(userID)
	logrus.Debugln(user)
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	contactListObject, er2 := contCont.ContactListService.ContactListObjectWithId(contactListInfoId)
	if er2 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListDoesNotExist)
	}

	//	if user.ID.Hex() != contactListObject.UserID.Hex() && !contactListObject.IsDoNotContactList {
	// given user is not authorized to make this call
	//		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListUnauthorizedForAction)
	//	}

	// contact list verification done
	// now upload contact list to s3
	ucl, err := contCont.generateCSVForContactList(*contactListObject)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	// Marshal provided interface into JSON structure
	sf, _ := json.Marshal(ucl)

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Success", "Success", sf)
}

// private method
func (contCont *ContactListController) generateCSVForContactList(contactList model.ContactList) (*model.ContactList, error) {

	dstPath, err := contCont.ContactListService.GenerateCSV(contactList)
	if err != nil {
		return nil, err
	}

	// upload file to s3
	tempPath := "/account/" + contactList.UserID.Hex() + "/contactlist/" + contactList.ID.Hex() + "/"
	s3FilePath, err := contCont.S3Service.Upload(tempPath, dstPath)
	if err != nil {
		// delete file from local
		cmlutils.DeleteFile(dstPath)
		return nil, err
	}

	// delete local file after uploading
	cmlutils.DeleteFile(dstPath)

	// update s3 Path
	if err := contCont.ContactListService.UpdateS3FilePath(&contactList, s3FilePath); err != nil {
		contCont.S3Service.DeleteS3Object(cmlutils.S3BucketName(), s3FilePath)
		logrus.Errorln("Error Contact list update : " + err.Error())
		return nil, err
	}

	return &contactList, nil
}

// private method
// this method process the sound file present at the given destination

//func (contCont *ContactListController) processUploadContactList(name string, filename string, userInfo *model.User, originalFilePath string, numberColumnValue int, isHeaderPresent int, contacttype int8, remDup bool, shufflenumber bool, scrubLandline bool, scrubCellPhone bool, areacode bool, scrubdnc bool) (*model.ContactList, int, error) {
//
//	// file path name
//	//fileBaseName := filepath.Base(originalFilePath)
//	//fileName := strings.TrimSuffix(fileBaseName, filepath.Ext(fileBaseName))
//
//	contactList := model.ContactList{}
//	contactList.ID = bson.NewObjectId()
//	contactList.FileName = filename
//	contactList.Name = name
//	contactList.Type = contacttype
//	contactList.RemoveDup = remDup
//	contactList.Shuffle = shufflenumber
//	contactList.ScrubLandLine = scrubLandline
//	contactList.ScrubCellPhone = scrubCellPhone
//	contactList.AreaCode = areacode
//	contactList.ScrubDNC = scrubdnc
//	destinationPath := "/account/" + userInfo.ID.Hex() + "/contactlist/" + contactList.ID.Hex() + "/"
//	s3FilePath, err := contCont.S3Service.Upload(destinationPath, originalFilePath)
//	if err != nil {
//		return nil, 0, err
//	}
//
//	contactList.FileS3Path = s3FilePath
//	contactList.UserID = userInfo.ID
//	//	contactList.IsDoNotContactList = false
//	contactList.Status = model.ContactListStatusProcessing
//	if err := contCont.ContactListService.SaveContactList(&contactList); err != nil {
//		// delete file from s3
//		contCont.S3Service.DeleteS3Object(cmlutils.S3BucketName(), s3FilePath)
//		logrus.Errorln("Error Save Contact list")
//		logrus.Errorln(contactList)
//		return nil, 0, err
//	}
//	go func() {
//		// extract numbers from the contact list
//		err, invalidnumber := contCont.extractNumbersFromUploadedContactList(userInfo, &contactList, originalFilePath, numberColumnValue, isHeaderPresent)
//		if err != nil {
//			contCont.contactListSaveFailOver(contactList, s3FilePath)
//			return
//		}
//
//		logrus.Infoln(invalidnumber)
//		// get user detail object
//
//		// update user object in db
//		if err := contCont.UserService.SaveUser(userInfo); err != nil {
//			contCont.contactListSaveFailOver(contactList, s3FilePath)
//			//return nil, 0, err
//			return
//		}
//		//delete locally saved file
//		cmlutils.DeleteFile(originalFilePath)
//	}()
//	return &contactList, 0, nil
//}

func (contCont *ContactListController) contactListSaveFailOver(contactList model.ContactList, s3FilePath string) {
	// delete already saved contact list
	contCont.ContactListService.DeleteContactList(&contactList)
	// delete file from s3
	contCont.S3Service.DeleteS3Object(cmlutils.S3BucketName(), s3FilePath)
}

func (contCont *ContactListController) cronDNCScrub() {
	// TODO: some work here
}

// private method
// this method extracts numbers from the uploaded csv file
// TODO: Refactor this function
//func (contCont *ContactListController) extractNumbersFromUploadedContactList(userInfo *model.User, contactList *model.ContactList, originalFilePath string, numberColumnValue int, isHeaderPresent int) (error, int) {
//	maxNbConcurrentGoroutines := 5
//	concurrentGoroutines := make(chan struct{}, maxNbConcurrentGoroutines)
//	var wg sync.WaitGroup
//	var invalidnumber int
//	var numberslice []string
//	var flag bool
//	var duplicatenumbers int
//	var mobilenumber int
//	var landlinenumber int
//	var numbertype string
//	var validUSNumber bool
//	var last string
//	var dnccount int
//	// read file
//	file, err := os.Open(originalFilePath)
//	if err != nil {
//		return err, 0
//	}
//	defer file.Close()
//
//	r := csv.NewReader(file)
//	numberList := model.NumberList{}
//	r.FieldsPerRecord = -1
//	recordCount := 0
//	batchCount := 0
//	listCount := 0
//	if contactList.Type == model.ContactListTypeDNC {
//		contactList.AreaCode = false
//		contactList.ScrubLandLine = false
//		contactList.Shuffle = false
//		contactList.ScrubCellPhone = false
//		contactList.RemoveDup = true
//
//	}
//	if contactList.Type == model.ContactListTypeCallerGroups {
//		contactList.ScrubLandLine = false
//		contactList.ScrubCellPhone = false
//	}
//	fileName, err := cmlutils.RemoveNonAlphaNumeric(contactList.Name)
//	if err != nil {
//		fileName = "contacts"
//	}
//
//	filePath := fmt.Sprintf("%s%s.csv", cmlconstants.TempDestinationPath, fileName)
//	file1, err1 := os.Create(filePath)
//	if err1 != nil {
//		return nil, 0
//	}
//
//	writer := csv.NewWriter(file1)
//	defer func() {
//		cmlutils.DeleteFile(filePath)
//	}()
//	// declare variable to keep all the created contact group references
//	contactGroupsReferences := model.ContactGroupRefs{}
//	if rec, err := r.Read(); err != nil { //read header
//		return errors.New(cmlmessages.ContactListReadError), 0
//	} else {
//		if !(len(rec) > numberColumnValue) {
//			// column number value exceeds that of actual column
//			return errors.New("number column value exceeds that of actual columns present in file"), 0
//		}
//
//		if isHeaderPresent >= 1 {
//			// there is header present in first row
//			// skip that
//		} else {
//			number, err := extractNumber(rec, numberColumnValue)
//			if err == nil {
//				if contactList.Type == model.ContactListTypeCallerGroups {
//					if validateCallerNumber(number.Number) == true {
//						number.Number = strings.ToLower(number.Number)
//						number.Number = strings.ToLower(number.Number)
//						count := strings.Count(number.Number, "x")
//						last = replacenumber(number.Number, count)
//						if cmlutils.CallerIDVerification(last) == true {
//							validUSNumber = true
//						} else {
//							logrus.Infoln("invalid caller id", last)
//							validUSNumber = false
//						}
//					}
//				}
//				numberslice = append(numberslice, number.Number)
//				if cmlutils.IsUSNumber(number.Number) == true || validUSNumber == true {
//
//					// here extract last 10 digits from the number
//					number.Number = cmlutils.ExtractNumberFromUSNumber(number.Number)
//					firstthree := number.Number[0:3]
//					firstsix := number.Number[0:6]
//					wirelessnumber := contCont.Wirelessnumber(firstsix)
//					if wirelessnumber == true {
//						numbertype = "cellnumber"
//						mobilenumber++
//						recordCount++
//					} else {
//						lindline := contCont.Landlinenumber(firstsix)
//						if lindline == true {
//							numbertype = "Lindlinenumber"
//							landlinenumber++
//							recordCount++
//						} else {
//							logrus.Debugln("Invalid Type")
//							number.NumberTypeStr = "Invalid Type"
//							invalidnumber++
//							recordCount++
//						}
//					}
//					number.TimeZone = contCont.Timezone(firstthree)
//					number.NumberTypeStr = numbertype
//					if contactList.RemoveDup == true && contactList.Type == model.ContactListTypeDefault {
//						// now use the extracted number
//						if len(numberList) == 0 {
//							if contactList.ScrubCellPhone == true && numbertype == "cellnumber" {
//
//							} else if contactList.ScrubLandLine == true && numbertype == "Lindlinenumber" {
//
//							} else {
//								err := writer.Write([]string{number.Number})
//								if err != nil {
//									logrus.Errorln(err)
//								}
//								numberList = append(numberList, *number)
//								listCount++
//								batchCount++
//							}
//						} else {
//							for _, n := range numberList {
//								if n.Number == number.Number {
//									flag = false
//									break
//								} else {
//									flag = true
//								}
//							}
//							if flag == true {
//								if contactList.ScrubCellPhone == true && numbertype == "cellnumber" {
//
//								} else if contactList.ScrubLandLine == true && numbertype == "Lindlinenumber" {
//
//								} else {
//									err := writer.Write([]string{number.Number})
//									if err != nil {
//										logrus.Errorln(err)
//									}
//									numberList = append(numberList, *number)
//									listCount++
//									batchCount++
//								}
//							}
//						}
//					} else {
//						if contactList.ScrubCellPhone == true && numbertype == "cellnumber" {
//
//						} else if contactList.ScrubLandLine == true && numbertype == "Lindlinenumber" {
//
//						} else {
//							err := writer.Write([]string{number.Number})
//							if err != nil {
//								logrus.Errorln(err)
//							}
//							numberList = append(numberList, *number)
//							listCount++
//							batchCount++
//						}
//					}
//				} else {
//					invalidnumber++
//				}
//			}
//
//		}
//
//	}
//
//	for {
//		rec, err := r.Read()
//		if err != nil {
//			if err == io.EOF {
//				logrus.Infoln("EOF reached")
//				break
//			} else {
//				logrus.Infoln("Error occured")
//			}
//		} else {
//			number, err := extractNumber(rec, numberColumnValue)
//			if err == nil {
//				if contactList.Type == model.ContactListTypeCallerGroups {
//					if validateCallerNumber(number.Number) == true {
//						number.Number = strings.ToLower(number.Number)
//
//						last := strings.Replace(number.Number, "x", strconv.Itoa(rand.Intn(9)), -1)
//
//						number.Number = last
//					} else {
//						continue
//					}
//				}
//				numberslice = append(numberslice, number.Number)
//				if cmlutils.IsUSNumber(number.Number) == true {
//					// here extract last 10 digits from the number
//					number.Number = cmlutils.ExtractNumberFromUSNumber(number.Number)
//
//					firstthree := number.Number[0:3]
//					firstsix := number.Number[0:6]
//					wirelessnumber := contCont.Wirelessnumber(firstsix)
//					if wirelessnumber == true {
//
//						numbertype = "cellnumber"
//						mobilenumber++
//						recordCount++
//					} else {
//						lindline := contCont.Landlinenumber(firstsix)
//						if lindline == true {
//
//							numbertype = "Lindlinenumber"
//							landlinenumber++
//							recordCount++
//						} else {
//							logrus.Debugln("Invalid Type")
//							number.NumberTypeStr = "Invalid Type"
//							recordCount++
//						}
//					}
//					number.TimeZone = contCont.Timezone(firstthree)
//					number.NumberTypeStr = numbertype
//					if contactList.ScrubLandLine == true && number.NumberTypeStr == "Lindlinenumber" {
//						continue
//					}
//					if contactList.ScrubCellPhone == true && number.NumberTypeStr == "cellnumber" {
//
//						continue
//					}
//					if contactList.RemoveDup == true && contactList.Type == model.ContactListTypeDefault {
//						// now use the extracted number
//						if len(numberList) == 0 {
//							err := writer.Write([]string{number.Number})
//							if err != nil {
//								logrus.Errorln(err)
//							}
//							numberList = append(numberList, *number)
//							listCount++
//							batchCount++
//						} else {
//							for _, n := range numberList {
//								if n.Number == number.Number {
//									flag = false
//									//duplicatenumbers++
//									break
//								} else {
//									flag = true
//								}
//							}
//							if flag == true {
//								err := writer.Write([]string{number.Number})
//								if err != nil {
//									logrus.Errorln(err)
//								}
//								numberList = append(numberList, *number)
//								listCount++
//								batchCount++
//							}
//						}
//					} else {
//						err := writer.Write([]string{number.Number})
//						if err != nil {
//							logrus.Errorln(err)
//						}
//						numberList = append(numberList, *number)
//						listCount++
//						batchCount++
//					}
//
//				} else {
//					invalidnumber++
//				}
//			}
//			if contactList.ScrubDNC == true {
//				if batchCount == cmlconstants.TCPABatchSize {
//					wg.Add(1)
//					number := numberList
//
//					go func() {
//						defer wg.Done()
//						logrus.Errorln("job start")
//						concurrentGoroutines <- struct{}{}
//
//						jobid, err := contCont.TCPAService.IsNumberInTCPA(number)
//						if err != nil {
//							logrus.Errorln(err.Error())
//						}
//						logrus.Infoln("Job Id ===>", jobid)
//
//						// save running job in database
//						if err := contCont.DNCJobsService.SaveDncJobObject(&model.DNCJobs{
//							Status:        model.DNCScrubJobTypeProcessing,
//							JobId:         jobid,
//							TTSListId: contactList.ID,
//						}); err != nil {
//							// TODO: handle this error, if error save contact list with error
//							return
//						}
//
//						<-concurrentGoroutines
//						logrus.Errorln("job end")
//					}()
//					numberList = model.NumberList{}
//					batchCount = 0
//				}
//			} else {
//				if batchCount == cmlconstants.ConfigContactListBatchSize {
//					contactGroup := createContactGroupObject(contactList)
//					contactGroup.Numbers = numberList
//					contactGroup.TotalNumbers = cmlconstants.ConfigContactListBatchSize
//
//					if err := contCont.ContactGroupService.SaveContactGroup(contactGroup); err != nil {
//						return err, 0
//					}
//					// cache contact group info to attach in contact list later
//					numberList = model.NumberList{}
//					batchCount = 0
//					contactGroupRef := model.ContactGroupRef{ContactGroupId: contactGroup.ID}
//					contactGroupsReferences = append(contactGroupsReferences, contactGroupRef)
//				}
//			}
//		}
//
//	}
//	wg.Wait()
//
//	for _, n := range numberslice {
//		if !numberExists(n) {
//			addnumber(n)
//		} else {
//			duplicatenumbers++
//		}
//	}
//	for k, _ := range duplicatenumber {
//		delete(duplicatenumber, k)
//	}
//	if contactList.Shuffle == true && contactList.Type == model.ContactListTypeDefault {
//		rand.Seed(time.Now().UnixNano())
//		rand.Shuffle(len(numberList), func(i, j int) { numberList[i], numberList[j] = numberList[j], numberList[i] })
//	}
//	if len(numberList) > 0 {
//		if contactList.ScrubDNC == true {
//			jobid, err := contCont.TCPAService.IsNumberInTCPA(numberList)
//			if err != nil {
//				logrus.Errorln(err.Error())
//			}
//			logrus.Infoln("Job Id =====>", jobid)
//		} else {
//			contactGroup := createContactGroupObject(contactList)
//			contactGroup.Numbers = numberList
//			contactGroup.TotalNumbers = int32(len(numberList))
//
//			if err := contCont.ContactGroupService.SaveContactGroup(contactGroup); err != nil {
//				return err, 0
//			}
//
//			// cache contact group info to attach in contact list later
//			contactGroupRef := model.ContactGroupRef{ContactGroupId: contactGroup.ID}
//			contactGroupsReferences = append(contactGroupsReferences, contactGroupRef)
//		}
//	}
//
//	contactList.ContactGroups = contactGroupsReferences
//
//	contactList.Invalid = int64(invalidnumber)
//	contactList.Duplicate = int64(duplicatenumbers)
//	if contactList.ScrubLandLine != true {
//		contactList.LandLine = int64(landlinenumber)
//	}
//
//	if contactList.ScrubCellPhone != true {
//		contactList.CellPhone = int64(mobilenumber)
//	}
//
//	if contactList.Type == model.ContactListTypeDefault && contactList.RemoveDup == true {
//		contactList.Imported = int64(recordCount) + contactList.Duplicate + contactList.Invalid
//	} else {
//		contactList.Imported = int64(recordCount) + contactList.Invalid
//	}
//
//	if contactList.ScrubDNC == false {
//		contactList.NumberCount = int64(listCount)
//	}
//
//	defer writer.Flush()
// 	s3FilePath, err := contCont.uploadContactListFile(contactList.UserID, contactList.ID, filePath)
//	if err == nil {
//		contactList.FileName = fileName
//		contactList.FileS3Path = s3FilePath
//	}
//
//	if err := contCont.ContactListService.SaveContactList(contactList); err != nil {
//		return err, 0
//	}
//
//	return nil, invalidnumber
//}

// TODO: Delete
// private method
func extractNumber(sourceRecords []string, numberColumn int) (*model.Number, error) {
	for i := 0; i < len(sourceRecords); i++ {
		csvRecord := sourceRecords[i]

		if i == numberColumn {
			phoneNumber := cmlutils.SimplifyPhoneNumber(csvRecord)
			return &model.Number{ID: bson.NewObjectId(), Number: phoneNumber}, nil
		}
	}

	return nil, errors.New("No Number found")
}

// private method
func createContactGroupObject(contactList *model.ContactList) *model.ContactGroup {
	contactGroup := model.ContactGroup{}

	contactGroup.ID = bson.NewObjectId()
	contactGroup.ContactListId = contactList.ID
	contactGroup.Numbers = model.NumberList{}

	return &contactGroup
}

//func numberExists(number string) bool {
//	_, ok := duplicatenumber[number]
//	return ok
//}
//
//func addnumber(number string) {
//	duplicatenumber[number] = true
//}

func (contCont *ContactListController) Wirelessnumber(number string) bool {
	s := contCont.ContactListService.RedisClient.HExists("cellnumber", number)
	str := s.String()
	if strings.Contains(str, "true") {
		return true
	} else {
		return false
	}
}
func (contCont *ContactListController) Landlinenumber(number string) bool {
	s := contCont.ContactListService.RedisClient.HExists("landlinenumber", number)
	str := s.String()
	if strings.Contains(str, "true") {
		return true
	} else {
		return false
	}
}
func (contCont *ContactListController) Timezone(number string) string {
	s := contCont.ContactListService.RedisClient.HGet("timezone", number)
	if s.Err() != nil {
		return ""
	}
	str := s.String()
	timezone := strings.Split(str, ":")
	tz := strings.TrimSpace(timezone[1])
	logrus.Infoln(tz)
	return tz
}

func (conCont *ContactListController) uploadContactListFile(userId bson.ObjectId, contId bson.ObjectId, path string) (string, error) {
	destinationPath := "/account/" + userId.Hex() + "/contactlist/" + contId.Hex() + "/"
	s3FilePath, err := conCont.S3Service.Upload(destinationPath, path)
	if err != nil {
		return "", err
	}

	return s3FilePath, err
}
func replacenumber(number string, count int) string {
	for i := 0; i <= count; i++ {
		number = strings.Replace(number, "x", strconv.Itoa(rand.Intn(9)), 1)
	}
	return number
}
