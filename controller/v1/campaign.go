package controllerv1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/network/response"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlconstants"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlmessages"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

type CampaignController struct {
	Controller
	UserService         *service.UserService
	UserDetailService   *service.UserDetailService
	SoundFileService    *service.SoundFileInfoService
	CampaignService     *service.CampaignService
	S3Service           *service.S3Service
	ContactListService  *service.ContactListService
	TTSListService      *service.TTSListService
	ContactGroupService *service.ContactGroupService
	CDRService          *service.CdrService
	EventService 		*service.EventsService
}

type (
	UpdateCampaignName struct {
		Name string `json:"name"`
	}

	UpdateLimits struct {
		CPM      int `json:"cpm"`
		Transfer int `json:"transfer"`
	}

	ExportCampaignResponse struct {
		ExportFile string `json:"exportFile"`
	}

	TestCallRequest struct {
		Number string `json:"number" validate:"required"`
	}

	TestTTSCallRequest struct {
		Number string `json:"number" validate:"required"`
		Text   string `json:"text" validate:"required"`
	}
	RecycleCampaign struct {
		ContactListName string `json:"name"`
		CallStatus      int    `json:"call_status"`
	}
)

// this method creates a campaign
func (camCont *CampaignController) CreateCampaign(c echo.Context) error {
	var flag bool
	cc := c.(model.AppContext)
	// Grab user id for which file is being uploaded
	userObject, err := camCont.UserService.UserObject(cc.ID.Hex())
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	// Stub an user to be populated from the body
	campaignInfo := model.Campaign{}
	if err := c.Bind(&campaignInfo); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	// validate input request body
	if err := c.Validate(campaignInfo); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	if campaignInfo.CallCampaign.CallerId == "" && campaignInfo.CallCampaign.CallerGroupId.Hex() == "" {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "CallerId or CallerGroupId is compulsory")
	}
	if len(campaignInfo.CallCampaign.ScheduleSettings.TimeZones) == 0 {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.CampaignSaveFail, "Please select valid time zone")

	}
	number := campaignInfo.CallCampaign.CallerId
	flag = validateCallerNumber(number)
	if flag == false {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Invalid Caller ID")
	}

	// user id has already been validated so no need to catch error here
	campaignInfo.Status = cmlconstants.CampaignStatusNew
	campaignInfo.UserID = bson.ObjectIdHex(cc.ID.Hex())
	campaignInfo.ParentID = userObject.ParentID

	// validate if all sound files ids are valid
	if err := camCont.validation(&campaignInfo); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	if campaignInfo.CallCampaign.IsTTS {
		// contact list settings
		ttsList, err := camCont.TTSListService.GetTTSList(campaignInfo.UserID, campaignInfo.CallCampaign.TTSListId)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.TTSContactListDoesNotExist)
		}

		campaignInfo.Stats.Total = ttsList.NumberCount
	} else {
		// contact list settings
		contantList, err := camCont.ContactListService.ContactListObjectWithId(campaignInfo.CallCampaign.ContactListId.Hex())
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListDoesNotExist)
		}

		campaignInfo.Stats.Total = contantList.NumberCount
	}

	// save campaign info
	if err := camCont.CampaignService.SaveCampaign(&campaignInfo); err != nil {
		// TODO : campaign save has failed
		// may be user uploaded contact list as well, which is save already
		// So either delete that contact list (preferred)
		// log this issue as warning in the logs with proper information
		// so that it can be looked into.
		logrus.Infoln("\n")
		logrus.Infoln(campaignInfo)
		logrus.Infoln("\n campaign save failed" + err.Error())
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.CampaignSaveFail)
	}

	cmg, _ := json.Marshal(campaignInfo)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Campaign created successfully", "Campaign created successfully", cmg)
}

// this method creates a campaign
func (camCont *CampaignController) CreateTTSCampaign(c echo.Context) error {
	var flag bool
	cc := c.(model.AppContext)
	// Grab user id for which file is being uploaded
	userObject, err := camCont.UserService.UserObject(cc.ID.Hex())
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	// Stub an user to be populated from the body
	campaignInfo := model.Campaign{}
	if err := c.Bind(&campaignInfo); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	// validate input request body
	if err := c.Validate(campaignInfo); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	if campaignInfo.CallCampaign.CallerId == "" && campaignInfo.CallCampaign.CallerGroupId.Hex() == "" {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "CallerId or CallerGroupId is compulsory")
	}
	if len(campaignInfo.CallCampaign.ScheduleSettings.TimeZones) == 0 {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.CampaignSaveFail, "Please select valid time zone")

	}
	number := campaignInfo.CallCampaign.CallerId
	flag = validateCallerNumber(number)
	if flag == false {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Invalid Caller ID")
	}

	// user id has already been validated so no need to catch error here
	campaignInfo.Status = cmlconstants.CampaignStatusNew
	campaignInfo.UserID = bson.ObjectIdHex(cc.ID.Hex())
	campaignInfo.ParentID = userObject.ParentID

	// validate if all sound files ids are valid
	if err := camCont.validation(&campaignInfo); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	// contact list settings
	ttsList, err := camCont.TTSListService.GetTTSList(campaignInfo.UserID, campaignInfo.CallCampaign.TTSListId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListDoesNotExist)
	}

	campaignInfo.Stats.Total = ttsList.NumberCount


	// save campaign info
	if err := camCont.CampaignService.SaveCampaign(&campaignInfo); err != nil {
		// TODO : campaign save has failed
		// may be user uploaded contact list as well, which is save already
		// So either delete that contact list (preferred)
		// log this issue as warning in the logs with proper information
		// so that it can be looked into.
		logrus.Infoln("\n")
		logrus.Infoln(campaignInfo)
		logrus.Infoln("\n campaign save failed" + err.Error())
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.CampaignSaveFail)
	}

	cmg, _ := json.Marshal(campaignInfo)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Campaign created successfully", "Campaign created successfully", cmg)
}

func validateCallerNumber(number string) bool {
	var flag bool
	logrus.Infoln("B", number)
	number = strings.ToLower(number)
	logrus.Infoln("A", number)
	if number != "" {
		if strings.Contains(number, ",") {
			sArray := strings.Split(number, ",")
			for _, i := range sArray {
				callerid := strings.TrimSpace(i)
				if callerid == "" {
					continue
				}
				if strings.Contains(callerid, "x") {
					f2 := i[:2]
					if strings.Contains(f2, "x") == true {
						flag = false
						return flag
						//return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Invalid Caller ID")
					}
					count := strings.Count(callerid, "x")
					n := strings.Replace(callerid, "x", "4", count)
					flag = cmlutils.IsUSNumber(n)
				} else {
					flag = cmlutils.IsUSNumber(callerid)
				}
				if flag == true {
					continue
				} else {
					return flag
					//return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Invalid Caller ID")
					break
				}
			}
		} else if strings.Contains(number, "x") {
			number = strings.TrimSpace(number)
			f2 := number[:2]
			if strings.Contains(f2, "x") == true {
				flag = false
				return flag
				// return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Invalid Caller ID")
			}
			count := strings.Count(number, "x")
			n := strings.Replace(number, "x", "4", count)
			flag = cmlutils.IsUSNumber(n)
			return flag
			// return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Invalid Caller ID")

		} else {
			flag = cmlutils.IsUSNumber(number)
			return flag
		}
	}
	return true
}

func (camCont *CampaignController) validation(campaign *model.Campaign) error {
	if len(campaign.CallCampaign.ScheduleSettings.TimeZones) > 0 {
		for _, i := range campaign.CallCampaign.ScheduleSettings.TimeZones {
			if i == "" {
				continue
			}
			_, err := time.LoadLocation(i)
			if err != nil {
				return err
			}
		}
	}

	if !campaign.CallCampaign.IsTTS {
		contactlist, err := camCont.ContactListService.ContactListObjectWithId(campaign.CallCampaign.ContactListId.Hex())
		if err != nil {
			logrus.Infoln(err)
			return err
		}
		logrus.Infoln(contactlist)

		campaign.CallCampaign.ContactListName = contactlist.Name
	} else {
		ttslist, err := camCont.TTSListService.GetTTSList(campaign.UserID, campaign.CallCampaign.TTSListId)
		if err != nil {
			logrus.Infoln(err)
			return err
		}

		campaign.CallCampaign.ContactListName = ttslist.Name
	}

	if campaign.CallCampaign.DNCListId != "" {
		_, err := camCont.ContactListService.ContactListObjectWithId(campaign.CallCampaign.DNCListId.Hex())
		if err != nil {
			return err
		}
	} else {
		campaign.CallCampaign.DNCListId = ""
	}

	if campaign.CallCampaign.CallerGroupId != nil && campaign.CallCampaign.CallerGroupId.Valid() {
		_, err := camCont.ContactListService.ContactListObjectWithId(campaign.CallCampaign.CallerGroupId.Hex())
		if err != nil {
			return err
		}
	} else {
		campaign.CallCampaign.CallerGroupId = nil
	}

	if campaign.CallCampaign.SoundFileId != nil && campaign.CallCampaign.SoundFileId.Valid() {
		_, err := camCont.SoundFileService.SoundFileObjectWithId(campaign.CallCampaign.SoundFileId.Hex())
		if err != nil {
			return err
		}
	}

	return nil
}

func (camCont *CampaignController) UpdateCampaign(c echo.Context) error {
	// Grab user id for which file is being uploaded
	var flag bool
	cc := c.(model.AppContext)
	user, err := camCont.UserService.UserObject(cc.ID.Hex())
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	if user.IsLocked {
		return httpresponse.CreateBadResponse(&c, http.StatusUnauthorized, cmlmessages.PleaseTryAgain, cmlmessages.UserAccountLocked)
	}

	campaignId := c.Param("cid")

	oldcamp, err := camCont.CampaignService.CampaignObjectWithId(campaignId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusUnauthorized, cmlmessages.PleaseTryAgain, cmlmessages.CampaignDoesNotExist)
	}

	// Stub an user to be populated from the body
	campaignInfo := model.Campaign{}

	if err := c.Bind(&campaignInfo); err != nil {
		logrus.Infoln(err.Error())
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserRequestFormatIncorrect)
	}
	// validate input request body
	if err := c.Validate(campaignInfo); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	number := campaignInfo.CallCampaign.CallerId
	flag = validateCallerNumber(number)
	if flag == false {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Invalid Caller ID")
	}

	//campaignInfo.ID = oldcamp.ID
	campaignInfo.CreateDate = oldcamp.CreateDate
	campaignInfo.Name = oldcamp.Name
	campaignInfo.ID = oldcamp.ID
	campaignInfo.CallCampaign.ContactListId = oldcamp.CallCampaign.ContactListId
	campaignInfo.CallCampaign.TTSListId = oldcamp.CallCampaign.TTSListId
	campaignInfo.CallCampaign.DNCListId = oldcamp.CallCampaign.DNCListId
	campaignInfo.CallCampaign.ContactListName = oldcamp.CallCampaign.ContactListName
	campaignInfo.UserID = oldcamp.UserID
	campaignInfo.ParentID = oldcamp.ParentID
	campaignInfo.Status = oldcamp.Status
	campaignInfo.Stats = oldcamp.Stats
	campaignInfo.ListeningStats = oldcamp.ListeningStats
	campaignInfo.CallCampaign.ScheduleSettings.Id = oldcamp.CallCampaign.ScheduleSettings.Id

	if err := camCont.validation(&campaignInfo); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	if err := camCont.CampaignService.SaveCampaign(&campaignInfo); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Campaign name not updated.")
	}

	if campaignInfo.CallCampaign.ScheduleSettings.Id != "" {
		// send update campaign request to voip service
		//go func() {
		url := fmt.Sprintf(cmlutils.VOIPAPIPath()+"user/%s/campaign/%s/update", user.ID.Hex(), campaignInfo.ID.Hex())
		logrus.Debugln("Update URL : ", url)
		resp, err := http.Get(url)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Campaign not updated.")
		}
		defer resp.Body.Close()
		//}()
	}

	data, _ := json.Marshal(campaignInfo)

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Campaign  updated", "Campaign Successfully updated", data)
	//return nil
}

func (camCont *CampaignController) UpdateCampaignSpeed(c echo.Context) error {
	// Grab user id for which file is being uploaded
	var flag bool
	cc := c.(model.AppContext)
	user, err := camCont.UserService.UserObject(cc.ID.Hex())
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	if user.IsLocked {
		return httpresponse.CreateBadResponse(&c, http.StatusUnauthorized, cmlmessages.PleaseTryAgain, cmlmessages.UserAccountLocked)
	}

	campaignId := c.Param("cid")

	oldcamp, err := camCont.CampaignService.CampaignObjectWithId(campaignId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusUnauthorized, cmlmessages.PleaseTryAgain, cmlmessages.CampaignDoesNotExist)
	}

	// Stub an user to be populated from the body
	campaignInfo := model.CampaignSpeed{}

	if err := c.Bind(&campaignInfo); err != nil {
		logrus.Infoln(err.Error())
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserRequestFormatIncorrect)
	}
	// validate input request body
	if err := c.Validate(campaignInfo); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	number := oldcamp.CallCampaign.CallerId
	flag = validateCallerNumber(number)
	if flag == false {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Invalid Caller ID")
	}
	oldcamp.CallCampaign.Speed = campaignInfo.Speed
	//campaignInfo.ID = oldcamp.ID
	//campaignInfo.CreateDate = oldcamp.CreateDate
	//campaignInfo.Name = oldcamp.Name
	//campaignInfo.ID = oldcamp.ID
	//campaignInfo.CallCampaign.TTSListId = oldcamp.CallCampaign.TTSListId
	//campaignInfo.CallCampaign.DNCListId = oldcamp.CallCampaign.DNCListId
	//campaignInfo.UserID = oldcamp.UserID
	//campaignInfo.ParentID = oldcamp.ParentID
	//campaignInfo.Status = oldcamp.Status
	//campaignInfo.Stats = oldcamp.Stats
	//campaignInfo.ListeningStats = oldcamp.ListeningStats
	//campaignInfo.CallCampaign.ScheduleSettings.Id = oldcamp.CallCampaign.ScheduleSettings.Id

	if err := camCont.validation(oldcamp); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	if err := camCont.CampaignService.SaveCampaign(oldcamp); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Campaign name not updated.")
	}

	if oldcamp.CallCampaign.ScheduleSettings.Id != "" {
		// send update campaign request to voip service
		//go func() {
		url := fmt.Sprintf(cmlutils.VOIPAPIPath()+"user/%s/campaign/%s/update", user.ID.Hex(), oldcamp.ID.Hex())
		logrus.Debugln("Update URL : ", url)
		resp, err := http.Get(url)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Campaign not updated.")
		}
		defer resp.Body.Close()
		//}()
	}

	data, _ := json.Marshal(oldcamp)

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Campaign  updated", "Campaign Successfully updated", data)
	//return nil
}

func (camCont *CampaignController) Export(c echo.Context) error {
	actionName := c.QueryParam("action")

	userId := c.Param("id")
	user, err := camCont.UserService.UserObject(userId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	if user.IsLocked {
		return httpresponse.CreateBadResponse(&c, http.StatusUnauthorized, cmlmessages.PleaseTryAgain, cmlmessages.UserAccountLocked)
	}

	campaignId := c.Param("cid")

	url := fmt.Sprintf(cmlutils.CDRAPIPath()+"campaign/%s/export?action=%s", campaignId, actionName)
	logrus.Infoln("Campaign Export URL : ", url)
	resp, err := http.Get(url)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	defer resp.Body.Close()

	body, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
		return httpresponse.CreateRawResponse(&c, resp.StatusCode, body)
	}

	type ExportCampaignCdrResponse struct {
		ExportFile string `json:"data"`
	}

	var res ExportCampaignCdrResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	response := ExportCampaignResponse{ExportFile: cmlutils.S3FullPath() + res.ExportFile}
	data, _ := json.Marshal(response)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", data)
}

// this method update a campaign
func (camCont *CampaignController) Recyclecampaign(c echo.Context) error {
	// Grab user id for which file is being uploaded
	cc := c.(model.AppContext)
	campaignId := c.Param("cid")
	statusesStr := c.QueryParam("Status")
	str := strings.Split(statusesStr, ",")

	var statusesInt []int
	for _, i := range str {
		if strings.TrimSpace(i) == "" {
			continue
		}

		status, err := strconv.Atoi(i)
		if err != nil {
			continue
		}

		statusesInt = append(statusesInt, status)
	}

	user, err := camCont.UserService.UserObject(cc.ID.Hex())
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	if user.IsLocked {
		return httpresponse.CreateBadResponse(&c, http.StatusUnauthorized, cmlmessages.PleaseTryAgain, cmlmessages.UserAccountLocked)
	}

	camp, err := camCont.CampaignService.CampaignObjectWithId(campaignId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusUnauthorized, cmlmessages.PleaseTryAgain, cmlmessages.CampaignDoesNotExist)
	}

	body := RecycleCampaign{}
	if err := c.Bind(&body); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Please provide valid information")
	}
	contactLists := camp.CallCampaign.ContactListId

	go func() {
		var numbers []string
		for _, status := range statusesInt {
			if status == model.NotDialedCalls {
				numbs, err := camCont.GetUnDialedNumberForCampaign(camp.ID.Hex())
				if err != nil {
					continue
				}

				numbers = append(numbers, numbs...)
			} else {
				numbs, err := camCont.CDRService.CallStats(campaignId, contactLists, status)
				if err != nil {
					logrus.Errorln(fmt.Sprintf("Recycle Campaign :: Campaign Id : %s :: Error : %s", campaignId, err.Error()))
					continue
				}

				for _, num := range numbs {
					numbers = append(numbers, num.Number)
				}

			}
		}

		if len(numbers) == 0 {
			logrus.Errorln(fmt.Sprintf("Recycle Campaign :: Campaign Id : %s :: No Numbers Found ", campaignId))
			return
		}

		_, err := camCont.processRecycleCampaign(body, numbers, camp, cc.ID)
		if err != nil {
			logrus.Errorln(fmt.Sprintf("Recycle Campaign :: Campaign Id : %s :: Error : %s", campaignId, err.Error()))
			return
		}

	}()

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusCreated, "Campaign created successfully", " Recycle Campaign successfully")
}

// this method creates a campaign
func (camCont *CampaignController) DeleteCampaign(c echo.Context) error {
	//cc := c.(model.AppContext)

	campaignId := c.Param("cid")
	camp, err := camCont.CampaignService.CampaignObjectWithId(campaignId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.CampaignDoesNotExist)
	}

	if !camp.CanBeDeleted() {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.CampaignDeleteFail)
	}

	err = camCont.CampaignService.DeleteWithId(campaignId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	if camp.Status == cmlconstants.CampaignStatusStopped {
		// TODO: complete implementation
		// delete campaign run object as well
		// delete from redis as well
	}

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusCreated, "Success", "Successfully Deleted")
}

// this method creates a campaign
func (camCont *CampaignController) RerunCampaign(c echo.Context) error {
	// Grab user id for which file is being uploaded
	userId := c.Param("id")
	user, err := camCont.UserService.UserObject(userId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	if user.IsLocked {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserAccountLocked)
	}

	campaignId := c.Param("cid")
	camp, err := camCont.CampaignService.CampaignObjectWithId(campaignId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.CampaignDoesNotExist)
	}

	//	totalContacts := camp.CallCampaign.Stats.Total

	// Stub an user to be populated from the body
	newCamp := model.Campaign{}
	//	newCamp.Type = camp.Type
	newCamp.Status = cmlconstants.CampaignStatusNew
	newCamp.Name = camp.Name + "_rerun"
	newCamp.UserID = camp.UserID
	newCamp.ParentID = camp.ParentID
	//newCamp.BillingParentID = camp.BillingParentID
	//	newCamp.ContactList = camp.ContactList
	newCamp.CallCampaign = camp.CallCampaign
	//	newCamp.CallCampaign.Stats = model.Stats{}
	//	newCamp.CallCampaign.Stats.Total = totalContacts
	//	newCamp.CallCampaign.Schedule.StartDateUI = ""
	//	newCamp.CallCampaign.Schedule.StartDate = nil
	//	newCamp.CallCampaign.Schedule.StartTime = []int{}
	//	newCamp.CallCampaign.Schedule.IncludeDays = []int{}
	//	newCamp.CallCampaign.Schedule.ResumeNextDay = false

	// save campaign info
	if err := camCont.CampaignService.SaveCampaign(&newCamp); err != nil {
		// TODO : campaign save has failed
		// may be user uploaded contact list as well, which is save already
		// So either delete that contact list (preferred)
		// log this issue as warning in the logs with proper information
		// so that it can be looked into.
		logrus.Infoln("\n")
		logrus.Infoln(newCamp)
		logrus.Infoln("\n campaign save failed" + err.Error())
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.CampaignSaveFail)
	}

	cmg, _ := json.Marshal(newCamp)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Campaign created successfully", "Campaign created successfully", cmg)
}

// this method creates a campaign
func (camCont *CampaignController) UpdateLimits(c echo.Context) error {

	// Grab user id for which file is being uploaded
	userId := c.Param("id")
	user, err := camCont.UserService.UserObject(userId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	if user.IsLocked {
		return httpresponse.CreateBadResponse(&c, http.StatusUnauthorized, cmlmessages.PleaseTryAgain, cmlmessages.UserAccountLocked)
	}

	campaignId := c.Param("cid")
	camp, err := camCont.CampaignService.CampaignObjectWithId(campaignId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.CampaignDoesNotExist)
	}

	body := UpdateLimits{}
	if err := c.Bind(&body); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Please provide valid information")
	}

	if body.CPM > 250 {
		body.CPM = 250
	}

	if body.Transfer > 100 {
		body.Transfer = 100
	}

	if err := camCont.CampaignService.UpdateLimits(campaignId, body.CPM, body.Transfer); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Campaign limits not updated.")
	}

	data, _ := json.Marshal(camp)

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Campaign limits updated", "Campaign limits updated", data)
}

// this method can be used to start or stop campaign
func (camCont *CampaignController) CampaignDetail(c echo.Context) error {
	cc := c.(model.AppContext)

	// Grab user id for which file is being uploaded
	user, err := camCont.UserService.UserObject(cc.ID.Hex())
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	if user.IsLocked {
		return httpresponse.CreateBadResponse(&c, http.StatusUnauthorized, cmlmessages.PleaseTryAgain, cmlmessages.UserAccountLocked)
	}

	campaignId := c.Param("cid")
	camp, err := camCont.CampaignService.CampaignObjectWithId(campaignId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.CampaignDoesNotExist)
	}

	data, _ := json.Marshal(camp)

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Campaign Info", "Campaign Info", data)
}

// this method can be used to start or stop campaign
func (camCont *CampaignController) CampaignAction(c echo.Context) error {
	cc := c.(model.AppContext)
	actionName := c.QueryParam("action")
	//
	campaignId := c.Param("cid")

	url := fmt.Sprintf(cmlutils.VOIPAPIPath()+"user/%s/campaign/%s?action=%s", cc.ID.Hex(), campaignId, actionName)
	resp, err := http.Get(url)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	//
	defer resp.Body.Close()
	//
	body, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	logrus.Infoln(string(body))
	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Success", "Success")
	//return httpresponse.CreateRawResponse(&c, resp.StatusCode, body)
}

// this method actually routes the request to voip server api to send a test campaign call
func (camCont *CampaignController) TestCampaign(c echo.Context) error {
	cc := c.(model.AppContext)
	campaignId := c.Param("cid")

	body := TestCallRequest{}
	if err := c.Bind(&body); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Please provide valid information")
	}

	// validate input request body
	if err := c.Validate(body); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	byter, err := json.Marshal(body)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	url := fmt.Sprintf(cmlutils.VOIPAPIPath()+"user/%s/campaign/%s/test", cc.ID.Hex(), campaignId)
	logrus.Infoln("URL : ", url)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(byter))
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Success", "Success")
}

func (camCont *CampaignController) TestTTSCampaign(c echo.Context) error {
	cc := c.(model.AppContext)
	campaignId := c.Param("cid")

	body := TestTTSCallRequest{}
	if err := c.Bind(&body); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Please provide valid information")
	}

	// validate input request body
	if err := c.Validate(body); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	byter, err := json.Marshal(body)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	url := fmt.Sprintf(cmlutils.VOIPAPIPath()+"user/%s/campaign/%s/test/tts", cc.ID.Hex(), campaignId)
	logrus.Infoln("URL : ", url)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(byter))
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Success", "Success")
}

// this method actually routes the request to voip server api to send a test campaign call
func (camCont *CampaignController) RecycleCampaign(c echo.Context) error {
	cc := c.(model.AppContext)
	campaignId := c.Param("cid")
	body := RecycleCampaign{}
	if err := c.Bind(&body); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Please provide valid information")
	}

	logrus.Infoln(campaignId)
	logrus.Infoln(cc.ID)
	logrus.Infoln(body)

	campaigninfo, err1 := camCont.CampaignService.CampaignObjectWithCampaignIdAndUserId(campaignId, cc.ID.Hex())
	if err1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err1.Error())
	}
	logrus.Infoln("campaigninfo", campaigninfo)
	if body.CallStatus == 1 {

	}

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Success", "Success")
}

// get list of all contact lists
// this method also supports server side pagination
func (camCont *CampaignController) GetCampaignStat(c echo.Context) error {

	// Grab user id for which file is being uploaded
	campaignInfo := model.CompareCampaign{}

	if err := c.Bind(&campaignInfo); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	//campaignList, err := camCont.CampaignService.CampaignListStat(campaignInfo.Ids)
	//if err != nil {
	//	return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	//}
	type CampaignStat struct {
		ID       bson.ObjectId `json:"id"`
		Live     int           `json:"live"`
		Transfer int           `json:"transfer"`
	}

	var stats []CampaignStat

	for _, id := range campaignInfo.Ids {
		if id == "" || !bson.IsObjectIdHex(id) {
			continue
		}

		stat := camCont.EventService.GetCampaignStats(id)

		stats = append(stats, CampaignStat{
			ID:       bson.ObjectIdHex(id),
			Live:     stat.RingingCount + stat.AnswerCount,
			Transfer: stat.TransferCount,
		})
	}

	//logrus.Infoln(campaignList)

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(stats)

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Campaign Stats fetched successfully", "Campaign Stats fetched successfully", data)
}

func (camCont *CampaignController) GetCampaignListForUser(c echo.Context) error {
	cc := c.(model.AppContext)

	// Grab user id for which file is being uploaded
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")

	pageNumber, errPage := strconv.Atoi(page)
	limitNumber, errLimit := strconv.Atoi(limit)

	if errPage != nil || errLimit != nil {
		pageNumber = 1
		limitNumber = 0
	}

	campaignList, count, err := camCont.CampaignService.CampaignListListingWithUserId(cc.ID.Hex(), pageNumber, limitNumber)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	var result model.CampaingList
	for _, c := range campaignList {
		stats := camCont.EventService.GetCampaignStats(c.ID.Hex())
		c.Stats.LiveAleg = int64(stats.AnswerCount + stats.RingingCount)
		c.Stats.LiveTransfer = int64(stats.TransferCount)
		result = append(result, c)
	}

	pageInfo := model.Page{}
	pageInfo.Page = pageNumber
	pageInfo.Limit = limitNumber
	pageInfo.TotalCount = count

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(result)
	pageData, _ := json.Marshal(pageInfo)

	return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusOK, "Campaign List fetched successfully", "Campaign List fetched successfully", data, pageData)
}

func (camCont *CampaignController) DuplicateCampaign(c echo.Context) error {

	cc := c.(model.AppContext)
	campaignId := c.Param("cid")
	logrus.Infoln(campaignId)
	duplicateobject := new(model.DuplicateCampaign)
	if err := c.Bind(duplicateobject); err != nil {
		return err
	}
	// validate input request body
	if err := c.Validate(duplicateobject); err != nil {
		logrus.Infoln(err)
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	logrus.Infoln(cc.ID.String())

	camp, err := camCont.CampaignService.CampaignObjectWithId(campaignId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.CampaignDoesNotExist)
	}

	if camp.CallCampaign.IsTTS {
		ttsid := bson.ObjectIdHex(duplicateobject.TTSListID)
		ttslist, err := camCont.TTSListService.GetTTSList(cc.ID, ttsid)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListDoesNotExist)
		}

		logrus.Infoln(ttslist)
		logrus.Infoln(ttslist.NumberCount)
		total := ttslist.NumberCount
		logrus.Infoln(camp)
		stats := model.Stats{}
		listenstate := model.ListeningStats{}
		camp.Stats = stats
		camp.ListeningStats = listenstate
		camp.Stats.Total = total
		camp.Name = duplicateobject.Name
		camp.CallCampaign.ContactListName = ttslist.Name
		camp.CallCampaign.TTSListId = ttslist.ID
		camp.Status = 1
		camp.ID = bson.NewObjectId()
		camp.Cost.CampaignCost = 0
		if err := camCont.CampaignService.SaveCampaign(camp); err != nil {
			// TODO : campaign save has failed
			// may be user uploaded contact list as well, which is save already
			// So either delete that contact list (preferred)
			// log this issue as warning in the logs with proper information
			// so that it can be looked into.
			logrus.Infoln("\n")
			logrus.Infoln(camp)
			logrus.Infoln("\n campaign save failed" + err.Error())
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.CampaignSaveFail)
		}

		logrus.Infoln(camp)
		cmg, _ := json.Marshal(camp)
		return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Campaign created successfully", " Duplicate Campaign successfully", cmg)

	}
	contactlist, err := camCont.ContactListService.ContactListObjectWithIdAndUserId(duplicateobject.ContactListId, cc.ID)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListDoesNotExist)
	}
	logrus.Infoln(contactlist)
	logrus.Infoln(contactlist.NumberCount)
	total := contactlist.NumberCount
	logrus.Infoln(camp)
	stats := model.Stats{}
	listenstate := model.ListeningStats{}
	camp.Stats = stats
	camp.ListeningStats = listenstate
	camp.Stats.Total = total
	camp.Name = duplicateobject.Name
	camp.CallCampaign.ContactListName = contactlist.Name
	camp.CallCampaign.ContactListId = contactlist.ID
	camp.Status = 1
	camp.ID = bson.NewObjectId()
	camp.Cost.CampaignCost = 0
	if err := camCont.CampaignService.SaveCampaign(camp); err != nil {
		// TODO : campaign save has failed
		// may be user uploaded contact list as well, which is save already
		// So either delete that contact list (preferred)
		// log this issue as warning in the logs with proper information
		// so that it can be looked into.
		logrus.Infoln("\n")
		logrus.Infoln(camp)
		logrus.Infoln("\n campaign save failed" + err.Error())
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.CampaignSaveFail)
	}

	logrus.Infoln(camp)
	cmg, _ := json.Marshal(camp)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Campaign created successfully", " Duplicate Campaign successfully", cmg)

}
func (camCont *CampaignController) Wirelessnumber(number string) bool {
	s := camCont.ContactListService.RedisClient.HExists("cellnumber", number)
	str := s.String()
	if strings.Contains(str, "true") {
		return true
	} else {
		return false
	}
}
func (camCont *CampaignController) Landlinenumber(number string) bool {
	s := camCont.ContactListService.RedisClient.HExists("landlinenumber", number)
	str := s.String()
	if strings.Contains(str, "true") {
		return true
	} else {
		return false
	}
}
func (camCont *CampaignController) Timezone(number string) string {
	s := camCont.ContactListService.RedisClient.HGet("timezone", number)
	str := s.String()
	timezone := strings.Split(str, ":")
	tz := strings.TrimSpace(timezone[1])
	return tz
}

func (camCont *CampaignController) uploadContactListFile(userId bson.ObjectId, contId bson.ObjectId, path string) (string, error) {
	destinationPath := "/account/" + userId.Hex() + "/contactlist/" + contId.Hex() + "/"
	s3FilePath, err := camCont.S3Service.Upload(destinationPath, path)
	if err != nil {
		return "", err
	}

	return s3FilePath, err
}
func (camCont *CampaignController) GetUnDialedNumberForCampaign(campaignId string) ([]string, error) {
	var numbers []string
	est := fmt.Sprintf("C_NIds_EST_%s", campaignId)
	cst := fmt.Sprintf("C_NIds_CST_%s", campaignId)
	mst := fmt.Sprintf("C_NIds_MST_%s", campaignId)
	pst := fmt.Sprintf("C_NIds_PST_%s", campaignId)

	zones := []string{
		est, cst, mst, pst,
	}

	for _, i := range zones {
		s, err := camCont.CampaignService.RedisClient.LRange(i, 0, -1).Result()
		if err != nil {
			return nil, err
		}
		numbers = append(numbers, s...)
	}

	if len(numbers) == 0 {
		return nil, errors.New("Numbers are not found aginst Call Status")
	}

	return numbers, nil
}

func (camCont *CampaignController) processRecycleCampaign(body RecycleCampaign, numbers []string, camp *model.Campaign, userID bson.ObjectId) ([]byte, error) {
	var campaign model.Campaign
	contactList := model.ContactList{}
	contactList.ID = bson.NewObjectId()
	contactList.FileName = body.ContactListName
	contactList.Name = body.ContactListName
	contactList.Type = model.ContactListTypeDefault
	contactList.RemoveDup = false
	contactList.Shuffle = false
	contactList.ScrubLandLine = false
	contactList.ScrubCellPhone = false
	contactList.UserID = userID

	fileName, err := cmlutils.RemoveNonAlphaNumeric(body.ContactListName)
	if err != nil {
		fileName = "contacts"
	}

	userinfo, err := camCont.UserService.UserObject(userID.Hex())
	if err != nil {
		logrus.Errorln(err.Error())
		return nil, err
	}

	filepath, err := camCont.ContactListService.NumbersToCSV(numbers, fileName)
	if err != nil {
		logrus.Errorln(err.Error())
		return nil, err
	}
	numbercol := 0
	headerPresent := false
	RemoveDuplicate := false
	ShuffleNumber := false
	ScrubLandline := false
	ScrubCellPhone := false
	AreaCodeDialing := false
	ScrubDNC := false
	Randomize := false
	cl, err := camCont.ContactListService.SaveContactListNew(&service.ContactListSaveOptions{Name: body.ContactListName,
		FileName:        fileName,
		FilePath:        filepath.Name(),
		UserInfo:        userinfo,
		NumberColumn:    int8(numbercol),
		HeaderPresent:   headerPresent,
		ContactListType: model.ContactListTypeDefault,
		RemoveDuplicate: RemoveDuplicate,
		ShuffleNumber:   ShuffleNumber,
		ScrubLandline:   ScrubLandline,
		ScrubCellPhone:  ScrubCellPhone,
		AreaCodeDialing: AreaCodeDialing,
		ScrubDNC:        ScrubDNC,
		Randomize:       Randomize})

	var clist *model.ContactList
	for {
		clist, err = camCont.ContactListService.ContactListObjectWithId(cl.ID.Hex())
		if err != nil {
			continue
		}
		if clist.Status == model.ContactListStatusActive {
			break
		} else {
			continue
		}
	}
	id := clist.ID
	total := clist.NumberCount
	campaign.ID = bson.NewObjectId()
	campaign.CallCampaign = camp.CallCampaign
	campaign.Stats = model.Stats{}
	campaign.Stats.Total = total
	campaign.ListeningStats = model.ListeningStats{}
	campaign.ParentID = camp.ParentID
	campaign.CallCampaign.ContactListName = contactList.Name
	campaign.Status = cmlconstants.CampaignStatusNew
	campaign.UserID = camp.UserID
	campaign.Cost = model.Cost{}
	campaign.Name = body.ContactListName
	campaign.CallCampaign.ContactListId = id
	err4 := camCont.CampaignService.SaveCampaign(&campaign)
	if err4 != nil {
		return nil, err4
	}
	uj, _ := json.Marshal(campaign)
	return uj, nil
}
