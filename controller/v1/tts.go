package controllerv1

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"github.com/gomarkho/sas-rvm-provapi/model"
	httpresponse "github.com/gomarkho/sas-rvm-provapi/network/response"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"strconv"
	"time"

	"github.com/gomarkho/sas-rvm-provapi/system/cmlmessages"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

type TTSController struct {
	TTSService         *service.TTSListService
	S3Service          *service.S3Service
	ContactListService *service.ContactListService
	UserService        *service.UserService
	CampaignService    *service.CampaignService
	Controller
}

// CreateUser creates a new user resource
func (smCon *TTSController) SaveTTSFile(c echo.Context) error {

	cc := c.(model.AppContext)

	name := c.FormValue("name")
	fileinfo, err := c.FormFile("file")
	if err != nil {

		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListSelectFile)
	}

	destinationPath := "/account/" + cc.ID.Hex() + "/ttsfile/" + bson.NewObjectId().Hex() + "/"

	dst, err1 := smCon.ContactListService.SaveTTSFile(fileinfo)
	if err1 != nil {
		return err1
	}
	s3, err := smCon.S3Service.Upload(destinationPath, dst)

	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, err.Error())
	}

	userid := cc.ID
	data := model.TTSFile{
		ID:          bson.NewObjectId(),
		UserID:      userid,
		S3Path:      s3,
		FileName:    name,
		CreatedDate: time.Now(),
	}
	file, err := smCon.TTSService.SaveTTSFile(data)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, err.Error())
	}
	cmlutils.DeleteFile(dst)
	datainfo, _ := json.Marshal(file)
	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "success", "success", datainfo)

}

// CreateUser creates a new user resource
func (smCon *TTSController) Import(c echo.Context) error {

	var req model.ImportData
	cc := c.(model.AppContext)
	if err := c.Bind(&req); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}

	req.UserID = cc.ID.Hex()

	cl, err := smCon.TTSService.Import(req)

	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, err.Error())
	}
	datainfo, _ := json.Marshal(cl)
	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "success", "success", datainfo)
}

// CreateUser creates a new user resource
func (smCon *TTSController) TTSList(c echo.Context) error {

	cc := c.(model.AppContext)
	search := c.QueryParam("search")
	limit := c.QueryParam("limit")
	page := c.QueryParam("page")

	fetchAll := c.QueryParam("listAll")

	fetchAllBool, err := strconv.ParseBool(fetchAll)
	if err != nil {
		fetchAllBool = false
	}
	pageNumber, errPage := strconv.Atoi(page)
	limitNumber, errLimit := strconv.Atoi(limit)

	if errPage != nil || errLimit != nil {
		pageNumber = 1
		limitNumber = 10
	}

	list, count, err := smCon.TTSService.TTSList(&model.CustomFilter{UserId: cc.ID, Search: search, Page: pageNumber, Limit: limitNumber}, fetchAllBool)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}

	pageInfo := model.Page{}
	pageInfo.Page = pageNumber
	pageInfo.Limit = limitNumber
	pageInfo.TotalCount = count

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(list)
	pageData, _ := json.Marshal(pageInfo)
	// Marshal provided interface into JSON structure

	return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusCreated, "Success", "Success", data, pageData)

}
func (smCon *TTSController) DeatilTTSList(c echo.Context) error {

	cc := c.(model.AppContext)

	ttsFileInfoId := c.Param("tid")

	listid := bson.ObjectIdHex(ttsFileInfoId)
	ttslist, err := smCon.TTSService.GetTTSList(cc.ID, listid)

	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}

	datainfo, err := json.Marshal(ttslist)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "success", "success", datainfo)
}

func (smCon *TTSController) DeleteTTS(c echo.Context) error {

	cc := c.(model.AppContext)

	ttsFileInfoId := c.Param("tid")

	listid := bson.ObjectIdHex(ttsFileInfoId)
	ttslist, err := smCon.TTSService.GetTTSList(cc.ID, listid)

	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}

	err = smCon.TTSService.DeleteList(ttslist.ID)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "TTS File Deleted successfully", "TTS File Deleted successfully")

}

func (smCon *TTSController) DeleteMultipleTTSList(c echo.Context) error {

	cc := c.(model.AppContext)
	// Grab user id for which file is being uploaded
	//contactListInfoId := c.Param("cid")
	contactlistObjects := ContactListIDs{}
	var count int

	if err := c.Bind(&contactlistObjects); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	user, er1 := smCon.UserService.UserObject(cc.ID.Hex())
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	for _, i := range contactlistObjects.ID {
		ttslist := bson.ObjectIdHex(i)
		contactListObject, er2 := smCon.TTSService.GetTTSList(cc.ID, ttslist)
		if er2 != nil {
			logrus.Errorln("TTSList Object Not Found")
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListDoesNotExist)
		}

		if user.ID.Hex() != contactListObject.UserID.Hex() {
			// given user is not authorized to make this call
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListUnauthorizedForAction)
		}

		bool, err := smCon.CampaignService.IsTTSListAttached(cc.ID.Hex(), i)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
		}

		if bool {
			// contact list is attached to some campaign, you can not delete this contact list
			continue
			//return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListAttachedWithCamapign)
		}

		if err := smCon.S3Service.CheckIfObjectExists(cmlutils.S3BucketName(), contactListObject.FileS3Path); err != nil {
			// contact list doesn't exist on server
			if err := smCon.TTSService.DeleteList(contactListObject.ID); err != nil {
				logrus.Errorln("TTS list delete failed => TTS list Id : " + contactListObject.ID.Hex())
				return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListDeleteOperationFailed)
			}
			count++
			// update user object in db
			if err := smCon.UserService.SaveUser(user); err != nil {
				logrus.Errorln("User save failed while deleting tts list => User id : " + user.ID.Hex() + " => TTS List : " + contactListObject.ID.Hex())
			}
			continue
			//continue
			//return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Contact list deleted successfully", "Contact list Deleted successfully")
			//return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListDeleteFailForS3)
		}

		if err := smCon.S3Service.DeleteS3Object(cmlutils.S3BucketName(), contactListObject.FileS3Path); err != nil {
			// delete file from s3
			if err := smCon.TTSService.DeleteList(contactListObject.ID); err != nil {
				logrus.Errorln("TTS list delete failed => tts list Id : " + contactListObject.ID.Hex())
				return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListDeleteOperationFailed)
			}
			count++
			// update user object in db
			if err := smCon.UserService.SaveUser(user); err != nil {
				logrus.Errorln("User save failed while deleting tts list => User id : " + user.ID.Hex() + " => TTS List : " + contactListObject.ID.Hex())
			}
			continue
			//return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Contact list deleted successfully", "Contact list Deleted successfully")
			//	return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListDeleteFailForS3)
		}

		if err := smCon.TTSService.DeleteList(contactListObject.ID); err != nil {
			logrus.Errorln("TTS list delete failed => Contact list Id : " + contactListObject.ID.Hex())
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListDeleteOperationFailed)
		}
		count++
		// update user object in db
		if err := smCon.UserService.SaveUser(user); err != nil {
			logrus.Errorln("User save failed while deleting contact list => User id : " + user.ID.Hex() + " => Contact List : " + contactListObject.ID.Hex())
		}
	}
	submessage := fmt.Sprintf("%d TTS list Deleted successfully", count)

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "TTS list deleted successfully", submessage)

}
