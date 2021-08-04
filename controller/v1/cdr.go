package controllerv1

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/network/response"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlmessages"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"io/ioutil"
	"net/http"
	"strconv"
)

type CdrController struct {
	Controller
	UserService     *service.UserService
	CDRService      *service.CdrService
	CampaignService *service.CampaignService
}
type ExportCdr struct {
	Email string `json:"email"`
}

func (cdrCont *CdrController) CdrList(c echo.Context) error {

	cc := c.(model.AppContext)

	// Grab user id for which file is being uploaded
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	fromDate := c.QueryParam("from")
	toDate := c.QueryParam("to")
	callstatus := c.QueryParam("callstatus")
	status := c.QueryParam("amdstatus")
	Type := c.QueryParam("calltype")
	campaign := c.QueryParam("campaignId")
	pageNumber, errPage := strconv.Atoi(page)
	limitNumber, errLimit := strconv.Atoi(limit)
	callType, err := strconv.Atoi(Type)
	amdStatus, err := strconv.Atoi(status)
	callStatus, err := strconv.Atoi(callstatus)
	if errPage != nil || errLimit != nil {
		pageNumber = 1
		limitNumber = 10
	}

	logrus.Infoln(fromDate, toDate, campaign, callType, amdStatus, callStatus)
	cdrList, count, err := cdrCont.CDRService.CdrList(cc.ID.Hex(), fromDate, toDate, int8(amdStatus), int8(callStatus), callType, campaign, pageNumber, limitNumber)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	pageInfo := model.Page{}
	pageInfo.Page = pageNumber
	pageInfo.Limit = limitNumber
	pageInfo.TotalCount = count

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(cdrList)
	pageData, _ := json.Marshal(pageInfo)
	if len(cdrList) == 0 {
		return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusOK, "No Record Found", "No Record Found", data, pageData)
	}
	return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusOK, "Cdr List fetched successfully", "Cdr List fetched successfully", data, pageData)
}

func (cdrCont *CdrController) CdrDaily(c echo.Context) error {

	cc := c.(model.AppContext)

	// Grab user id for which file is being uploaded
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	fromDate := c.QueryParam("from")
	toDate := c.QueryParam("to")
	Type := c.QueryParam("calltype")
	campaign := c.QueryParam("campaignId")
	pageNumber, errPage := strconv.Atoi(page)
	limitNumber, errLimit := strconv.Atoi(limit)
	callType, err := strconv.Atoi(Type)
	if err != nil {
		logrus.Errorln(err.Error())
	}

	if errPage != nil || errLimit != nil {
		pageNumber = 1
		limitNumber = 10
	}

	tempLimit := 50000
	tempPage := 1
	results := map[string]model.DailyCDR{}
	for {
		cdrList, count, err := cdrCont.CDRService.CdrList(cc.ID.Hex(), fromDate, toDate, 0, 0, callType, campaign, tempPage, tempLimit)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
		}

		for _, i := range cdrList {
			// first extract day
			date := i.StartDate
			dateString := fmt.Sprintf("%d-%d-%d", date.Year(), date.Month(), date.Day())
			info, ok := results[dateString]
			if !ok {
				info = model.DailyCDR{}
			}

			info.Date = dateString
			info.BillDuration += i.BillDuration
			info.Call += 1
			if i.Disposition == model.DispositionAnswer {
				info.Answer += 1
			}
			info.IncrementDuration += i.BillingIncr
			if info.Call == 0 || info.Answer == 0 {
				info.ALOC = 0
			} else {
				info.ALOC = info.BillDuration / info.Answer
			}

			info.Cost += float64(i.Cost)

			results[dateString] = info

		}

		if count <= (tempPage * tempLimit) {
			break
		}

		tempPage += 1

	}

	dailyList := model.DailyList{}
	for _, i := range results {
		dailyList = append(dailyList, i)
	}

	pageInfo := model.Page{}
	pageInfo.Page = pageNumber
	pageInfo.Limit = limitNumber
	pageInfo.TotalCount = len(dailyList)

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(dailyList)
	pageData, _ := json.Marshal(pageInfo)
	if len(data) == 0 {
		return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusOK, "No Record Found", "No Record Found", data, pageData)
	}
	return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusOK, "Cdr List fetched successfully", "Cdr List fetched successfully", data, pageData)
}

func (cdrCont *CdrController) CdrHourlyAleg(c echo.Context) error {
	cc := c.(model.AppContext)

	// Grab user id for which file is being uploaded
	fromDate := c.QueryParam("from")
	toDate := c.QueryParam("to")
	callstatus := c.QueryParam("callstatus")
	status := c.QueryParam("amdstatus")
	campaign := c.QueryParam("campaignId")

	amdStatus, err := strconv.Atoi(status)
	if err != nil {
		amdStatus = 0
	}

	callStatus, err := strconv.Atoi(callstatus)
	if err != nil {
		callStatus = 0
	}

	tempLimit := 50000
	tempPage := 1
	results := map[string]model.HourlyCDR{}

	for {
		cdrList, count, err := cdrCont.CDRService.CdrList(cc.ID.Hex(), fromDate, toDate, int8(amdStatus), int8(callStatus), model.CallTypeALeg, campaign, tempPage, tempLimit)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
		}

		for _, i := range cdrList {
			// first extract day
			date := i.StartDate
			hourString := fmt.Sprintf("%d", date.Hour())
			info, ok := results[hourString]
			if !ok {
				info = model.HourlyCDR{}
			}

			info.From = date.Hour()
			info.To = date.Hour() + 1
			info.TotalCall += 1
			info.BillDuration += i.BillDuration
			info.IncrementDuration += i.BillingIncr

			if i.Disposition == model.DispositionAnswer {
				info.Answer += 1
			}

			if i.Disposition == model.DispositionBusy {
				info.Busy += 1
			}

			if i.Disposition == model.DispositionNoAnswer {
				info.NoAnswer += 1
			}

			if i.Disposition == model.DispositionError {
				info.Failed += 1
			}

			if i.AMDStatus == model.AMDStatusVM {
				info.VoiceMail += 1
			}
			if info.Answer == 0 {
				info.ALOC = 0
			} else {

				info.ALOC = info.BillDuration / info.Answer
			}
			if info.TotalCall == 0 {
				info.ASR = 0.0
			} else {
				info.ASR = float64(info.Answer) / float64(info.TotalCall)
				info.ASR = info.ASR * 100
			}

			info.Cost += float64(i.Cost)

			results[hourString] = info

		}

		if count <= (tempPage * tempLimit) {
			break
		}

		tempPage += 1

	}

	list := model.HourlyCdrList{}

	for _, i := range results {
		list = append(list, i)
	}
	data, _ := json.Marshal(list)
	if len(list) == 0 {
		return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "No Record Found", "No Record Found", data)
	}
	// Marshal provided interface into JSON structure

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Cdr List fetched successfully", "Cdr List fetched successfully", data)
}

func (cdrCont *CdrController) CdrHourlyBleg(c echo.Context) error {

	cc := c.(model.AppContext)
	fromDate := c.QueryParam("from")
	toDate := c.QueryParam("to")
	callstatus := c.QueryParam("callstatus")
	status := c.QueryParam("amdstatus")
	campaign := c.QueryParam("campaignId")
	amdStatus, err := strconv.Atoi(status)
	if err != nil {
		amdStatus = 0
	}

	callStatus, err := strconv.Atoi(callstatus)
	if err != nil {
		callStatus = 0
	}

	tempLimit := 10000
	tempPage := 1
	results := map[string]model.HourlyCDR{}

	for {
		cdrList, count, err := cdrCont.CDRService.CdrList(cc.ID.Hex(), fromDate, toDate, int8(amdStatus), int8(callStatus), model.CallTypeBLeg, campaign, tempPage, tempLimit)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
		}

		for _, i := range cdrList {
			// first extract day
			date := i.StartDate
			hourString := fmt.Sprintf("%d", date.Hour())
			info, ok := results[hourString]
			if !ok {
				info = model.HourlyCDR{}
			}
			info.From = date.Hour()
			info.To = date.Hour() + 1
			info.TotalCall += 1
			info.BillDuration += i.BillDuration
			info.IncrementDuration += i.BillingIncr

			if i.Disposition == model.DispositionAnswer {
				info.Answer += 1
			}

			if i.Disposition == model.DispositionBusy {
				info.Busy += 1
			}

			if i.Disposition == model.DispositionNoAnswer {
				info.NoAnswer += 1
			}

			if i.Disposition == model.DispositionError {
				info.Failed += 1
			}

			if i.AMDStatus == model.AMDStatusVM {
				info.VoiceMail += 1
			}
			if info.Answer == 0 {
				info.ALOC = 0.0
			} else {
				info.ALOC = info.BillDuration / (info.Answer)
			}
			if info.TotalCall == 0 {
				info.ASR = 0.0
			} else {
				info.ASR = float64(info.Answer) / float64(info.TotalCall)
				info.ASR = info.ASR * 100
			}
			info.Cost += float64(i.Cost)

			results[hourString] = info

		}

		if count <= (tempPage * tempLimit) {
			break
		}

		tempPage += 1

	}

	list := model.HourlyCdrList{}

	for _, i := range results {
		list = append(list, i)
	}
	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(list)
	if len(list) == 0 {
		return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "No Record Found", "No Record Found", data)
	}

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Cdr List fetched successfully", "Cdr List fetched successfully", data)
}

func (cdrCont *CdrController) CompareCampaign(c echo.Context) error {

	// Grab user id for which file is being uploaded
	campaignInfo := model.CompareCampaign{}

	if err := c.Bind(&campaignInfo); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	tempLimit := 100000
	tempPage := 1
	results := map[string]model.CampaignCDR{}

	for {
		cdrList, count, err := cdrCont.CDRService.CdrListForCampaigns(campaignInfo.Ids, tempPage, tempLimit)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
		}

		for _, i := range *cdrList {
			// first extract day
			campaignId := fmt.Sprintf("%s", i.ID.Hex())
			info, ok := results[campaignId]
			if !ok {
				info = model.CampaignCDR{}
				campaign, err := cdrCont.CampaignService.CampaignObjectWithId(campaignId)
				if err != nil {
					return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
				}

				info.Contact = int(campaign.Stats.Total)
				info.Dial = int(campaign.Stats.Dialed)
				info.Redial = 0
				info.Answer = int(campaign.Stats.Answer)
				info.Live = int(campaign.Stats.Live)
				info.Minutes = int(campaign.Stats.Minutes)
				info.NoAnswer = int(campaign.Stats.NoAnswer)
				info.Vm = int(campaign.Stats.VoiceMail)
				info.CampaignID = campaign.ID.Hex()
				info.CampaignName = campaign.Name
				info.Stats = campaign.Stats
				info.ListeningStats = campaign.ListeningStats
				info.Cost = campaign.Cost
				//human:=info.Live-info.Vm
				if info.Answer == 0 {
					info.ALOC = 0.0
				} else {

					info.ALOC = info.Minutes / (info.Answer)

				}
				if info.Contact == 0 {
					info.ASR = 0.0
				} else {
					info.ASR = float64(info.Answer) / float64(info.Dial)
					info.ASR = info.ASR * 100
				}
				if info.Dial == 0 {
					info.LivePercent = 0.0
				} else {
					info.LivePercent = float64(info.Live) / float64(info.Dial)
				}

				if info.Dial == 0 {
					info.VmPercent = 0.0
				} else {
					info.VmPercent = float64(info.Vm) / float64(info.Dial)
				}
				if info.Contact == 0 {
					info.Progress = 0.0
				} else {
					info.Progress = float64(info.Dial) / float64(info.Contact)
				}
				info.XAns = int(campaign.Stats.TransferAns)

				if campaign.Stats.Transfer == 0 {
					info.XAnsPerc = 0.0
				} else {
					info.XAnsPerc = float64(campaign.Stats.TransferAns) / float64(campaign.Stats.Transfer)
				}

				if campaign.Stats.Transfer == 0 {
					info.CPX = 0.0
				} else {
					info.CPX = (campaign.Cost.TransferCost) / float64(campaign.Stats.Transfer)
				}

			}
			if info.Answer == 0 {
				info.ALOC = 0.0
			} else {
				info.ALOC = info.Minutes / (info.Answer)
			}
			//	info.Minutes += i.BillDuration

			results[campaignId] = info
		}

		if count <= (tempPage * tempLimit) {
			break
		}

		tempPage += 1
	}

	list := []model.CampaignCDR{}

	for _, i := range results {
		list = append(list, i)
	}
	data, err := json.Marshal(list)
	if err != nil {
		logrus.Errorln(err.Error())
	}

	if len(list) == 0 {
		return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "No Record Found", "No Record Found", data)
	}
	// Marshal provided interface into JSON structure

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Cdr List fetched successfully", "Cdr List fetched successfully", data)
}

func (cdrCont *CdrController) CampaignHourly(c echo.Context) error {

	cc := c.(model.AppContext)

	// Grab user id for which file is being uploaded
	date := c.QueryParam("date")
	campaignid := c.QueryParam("compaignId")

	if date == "" || campaignid == "" {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Please Provide Date and CampaignId")
	}
	tempLimit := 10000
	tempPage := 1
	results := map[string]model.CampaignHourly{}
	_, err := cdrCont.CampaignService.CampaignObjectWithId(campaignid)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	for {
		cdrList, count, err := cdrCont.CDRService.CdrList(cc.ID.Hex(), date, "", 0, 0, 0, campaignid, tempPage, tempLimit)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
		}

		for _, i := range cdrList {
			// first extract day
			date := i.StartDate
			hourString := fmt.Sprintf("%d", date.Hour())
			info, ok := results[hourString]
			if !ok {
				info = model.CampaignHourly{}
			}

			info.From = date.Hour()
			info.To = date.Hour() + 1
			info.CampaignID = i.CampaignId.Hex()
			info.Call += 1
			info.Minute += i.BillDuration
			info.Cost += float64(i.Cost)
			if i.CallType == model.CallTypeBLeg {
				info.XFer += 1
				info.XFerTC += float64(i.Cost)
				info.CPX = info.XFerTC / float64(info.XFer) // check formula with bilal
				info.XMin += i.BillDuration
			}

			if i.Disposition == model.DispositionAnswer {
				info.Answer += 1
			}

			if i.Disposition == model.DispositionDNC {
				info.DNC += 1
			}

			if i.Disposition == model.DispositionBusy {
				info.Busy += 1
			}

			if i.Disposition == model.DispositionNoAnswer {
				info.NoAns += 1
			}

			if i.Disposition == model.DispositionError {
				info.Failed += 1
			}

			if i.AMDStatus == model.AMDStatusVM {
				info.VM += 1
			}

			if i.CallType == model.CallTypeBLeg && i.Disposition == model.DispositionAnswer {
				info.XAns += 1
			}
			if info.Answer == 0 {
				info.ALOC = 0.0
			} else {
				info.ALOC = info.Minute / (info.Answer)
			}

			if info.Call == 0 {
				info.ASR = 0.0
			} else {
				info.ASR = float64(info.Answer) / float64(info.Call)
				info.ASR = info.ASR * 100
			}
			if info.XAns == 0 {
				info.XALOC = 0
			} else {
				info.XALOC = info.XMin / (info.XAns)
			}
			if info.XFer == 0 {
				info.XASR = 0.0
			} else {
				info.XASR = float64(info.XAns) / float64(info.XFer)
				info.XASR = info.XASR * 100
			}
			results[hourString] = info

		}

		if count <= (tempPage * tempLimit) {
			break
		}

		tempPage += 1
	}

	list := model.CampaignHourlyList{}

	for _, i := range results {
		list = append(list, i)
	}

	data, _ := json.Marshal(list)
	if len(list) == 0 {
		return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "No Record Found", "No Record Found", data)
	}
	// Marshal provided interface into JSON structure

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Cdr List fetched successfully", "Cdr List fetched successfully", data)
}

func (cdrCont *CdrController) ExportCdr(c echo.Context) error {
	// Grab user id for which file is being uploaded
	email := ExportCdr{}
	cc := c.(model.AppContext)
	fromDate := c.QueryParam("from")
	toDate := c.QueryParam("to")
	callstatus := c.QueryParam("callstatus")
	status := c.QueryParam("amdstatus")
	Type := c.QueryParam("calltype")
	campaign := c.QueryParam("campaignId")
	callType, _ := strconv.Atoi(Type)
	amdStatus, _ := strconv.Atoi(status)
	callStatus, _ := strconv.Atoi(callstatus)

	userinformation, err1 := cdrCont.UserService.UserObject(cc.ID.Hex())
	if err1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err1.Error())
	}
	parentinformation, err2 := cdrCont.UserService.UserObject(userinformation.ParentID.Hex())
	if err2 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err2.Error())
	}

	url := fmt.Sprintf(cmlutils.CDRAPIPath()+"export?parentId=%s&userId=%s&from=%s&to=%s&amdstatus=%d&callstatus=%d&campaignId=%s&calltype=%d", parentinformation.ID.Hex(), userinformation.ID.Hex(), fromDate, toDate, amdStatus, callStatus, campaign, callType)
	if err := c.Bind(&email); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

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
	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Cdr List export successfully", "You will receive an email in few minutes")
}

func (cdrCont *CdrController) CampaignDaily(c echo.Context) error {

	cc := c.(model.AppContext)

	// Grab user id for which file is being uploaded
	date := c.QueryParam("date")
	campaignid := c.QueryParam("compaignId")

	if date == "" || campaignid == "" {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Please Provide Date and CampaignId")
	}
	tempLimit := 10000
	tempPage := 1
	results := map[string]model.CampaignDaily{}

	campaign, err := cdrCont.CampaignService.CampaignObjectWithId(campaignid)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	for {
		cdrList, count, err := cdrCont.CDRService.CdrList(cc.ID.Hex(), date, "", 0, 0, 0, campaignid, tempPage, tempLimit)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
		}

		for _, i := range cdrList {
			// first extract day
			date := i.StartDate
			dateString := fmt.Sprintf("%d-%d-%d", date.Year(), date.Month(), date.Day())
			info, ok := results[dateString]
			if !ok {
				info = model.CampaignDaily{}
			}

			info.Date = dateString
			info.CampaignID = i.CampaignId.Hex()
			info.CampaignName = i.CampaignName
			info.Created = campaign.CreateDate.String()
			info.Calls += 1
			info.Minute += i.BillDuration
			info.Cost += float64(i.Cost)

			if i.CallType == model.CallTypeBLeg {
				info.Xfer += 1
				info.XferCost += float64(i.Cost)
				info.XferMin += i.BillDuration
				info.CPX = info.XferCost / float64(info.Xfer) // check formula with bilal
			}

			if campaign.Stats.Total == 0 {
				info.Progress = 0.0
			} else {
				info.Progress = float64(info.Calls) / float64(campaign.Stats.Total)
			}
			results[dateString] = info

		}

		if count <= (tempPage * tempLimit) {
			break
		}

		tempPage += 1
	}

	list := model.CampaignDailyList{}

	for _, i := range results {
		list = append(list, i)
	}

	data, err := json.Marshal(list)
	if err != nil {
		logrus.Errorln(err.Error())
	}

	if len(list) == 0 {
		return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "No Record Found", "No Record Found", data)
	}
	// Marshal provided interface into JSON structure

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Cdr List fetched successfully", "Cdr List fetched successfully", data)
}
