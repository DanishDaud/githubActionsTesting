package controllerv1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"github.com/gomarkho/sas-rvm-provapi/model"
	httpresponse "github.com/gomarkho/sas-rvm-provapi/network/response"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlmessages"
	"gopkg.in/mgo.v2/bson"
)

type NumbersCarrierGroupController struct {
	Controller
	//	UserRoleService      *service.UserRoleService
	NumbersCarrierGroupService *service.NumbersCarrierGroupService
}

type NumberCarrierGroup struct {
	NumberID      []string      `json:"numberID"`
	UserID        bson.ObjectId `bson:"userID"`
	ContactListID bson.ObjectId `json:"contactListID" bson:"contactListID"`
}

//// CreateUser creates a  ̰new user resource
func (uc *NumbersCarrierGroupController) CreateSingle(c echo.Context) error {
	var req NumberCarrierGroup
	cc := c.(model.AppContext)
	if err := c.Bind(&req); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}
	req.UserID = cc.ID
	id := c.Param("id")
	result, err := uc.NumbersCarrierGroupService.CreateSingle(service.TelnyxNumbersGroupCreateSingle{
		UserID:        req.UserID,
		NumberID:      id,
		ContactListID: req.ContactListID,
	})
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}
	uj, _ := json.Marshal(result)

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Number added to contactlist.", "Please resend account verification email", uj)
}

func (uc *NumbersCarrierGroupController) CreateMultiple(c echo.Context) error {

	var req NumberCarrierGroup
	cc := c.(model.AppContext)
	if err := c.Bind(&req); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}
	req.UserID = cc.ID
	result, err := uc.NumbersCarrierGroupService.CreateMultiple(service.TelnyxNumbersGroupCreateMultiple{
		UserID:        req.UserID,
		NumberID:      req.NumberID,
		ContactListID: req.ContactListID,
	})
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}
	uj, _ := json.Marshal(result)
	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Numbers added to contact list.", "Please resend account verification email", uj)
}

func (uc *NumbersCarrierGroupController) GetAll(c echo.Context) error {

	cc := c.(model.AppContext)

	// Grab user id for which file is being uploaded
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	search := c.QueryParam("search")
	id := c.QueryParam("id")

	var op model.GetAllTelnyxGroupRequest
	pageNumber, errPage := strconv.Atoi(page)
	if errPage != nil {
		pageNumber = 1
	}

	limitNumber, errLimit := strconv.Atoi(limit)
	if errLimit != nil {
		limitNumber = 10
	}
	if limitNumber < 10 {
		limitNumber = 10
	}

	if id == "" {
		op.UserID = cc.ID
		op.Limit = limitNumber
		op.Page = pageNumber
		op.Search = search

	} else {
		op.NumberID = bson.ObjectIdHex(id)
		op.UserID = cc.ID
		op.Limit = limitNumber
		op.Page = pageNumber
		op.Search = search
	}

	campaignList, count, err := uc.NumbersCarrierGroupService.GetAll(op)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	pageInfo := model.Page{}
	pageInfo.Page = pageNumber
	pageInfo.Limit = limitNumber
	pageInfo.TotalCount = count

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(campaignList)
	pageData, _ := json.Marshal(pageInfo)

	return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusOK, "Number List fetched successfully", "Number List fetched successfully", data, pageData)
}

func (uc *NumbersCarrierGroupController) DeleteMultiple(c echo.Context) error {

	var req service.TelnyxNumbersGroupDeleteMultipleRequest
	cc := c.(model.AppContext)
	if err := c.Bind(&req); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}
	req.UserID = cc.ID
	err := uc.NumbersCarrierGroupService.DeleteMultiple(service.TelnyxNumbersGroupDeleteMultipleRequest{
		UserID:        req.UserID,
		NumberID:      req.NumberID,
		ContactListID: req.ContactListID,
	})
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Number Removed From Contact Lists.", "Please resend account verification email")
}

func (uc *NumbersCarrierGroupController) DeleteSingle(c echo.Context) error {

	var req NumberCarrierGroup
	cc := c.(model.AppContext)
	if err := c.Bind(&req); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}
	req.UserID = cc.ID
	numberID := c.Param("id")
	err := uc.NumbersCarrierGroupService.DeleteSingle(service.TelnyxNumbersGroupDeleteSingleRequest{
		UserID:        req.UserID,
		NumberID:      numberID,
		ContactListID: req.ContactListID,
	})
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Number Removed From Contact List.", "Please resend account verification email")
}

func (uc *NumbersCarrierGroupController) GetContactListStats(c echo.Context) error {

	cc := c.(model.AppContext)
	// Grab user id for which file is being uploaded
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	search := c.QueryParam("search")
	id := c.QueryParam("id")

	var op model.GetContactListStatsRequest
	pageNumber, errPage := strconv.Atoi(page)
	if errPage != nil {
		pageNumber = 1
	}

	limitNumber, errLimit := strconv.Atoi(limit)
	if errLimit != nil {
		limitNumber = 10
	}
	if limitNumber < 10 {
		limitNumber = 10
	}

	if id == "" {
		op.UserID = cc.ID
		op.Limit = limitNumber
		op.Page = pageNumber
		op.Search = search

	} else {
		op.ContactListID = bson.ObjectIdHex(id)
		op.UserID = cc.ID
		op.Limit = limitNumber
		op.Page = pageNumber
		op.Search = search
	}

	campaignList, count, err := uc.NumbersCarrierGroupService.GetContactListStats(op)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	pageInfo := model.Page{}
	pageInfo.Page = pageNumber
	pageInfo.Limit = limitNumber
	pageInfo.TotalCount = count

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(campaignList)
	pageData, _ := json.Marshal(pageInfo)

	return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusOK, "Number List fetched successfully", "Number List fetched successfully", data, pageData)
}

//
//
//type TelnyxNumbersGroupRequest struct {
//	Numbers []string `json:"phone_numbers"`
//	UserID     bson.ObjectId `bson:"userid"`
//	TTSListID bson.ObjectId `json:"contactListID" bson:"contactListID"`
//}
//
//
//
//
//
//
//func (uc *TelnyxNumbersGroupController) PurchaseNumbers(c echo.Context) error {
//
//	// Stub an user to be populated from the body
//	var req ReqNumbers
//	cc := c.(model.AppContext)
//	if err := c.Bind(&req); err != nil {
//		return err
//	}
//	req.UserID = cc.ID
//	//phonenumbers := c.Param("phone_numbers")
//	result,err := uc.TelnyxService.PurchaseNumbers(service.PurchaseNumbersRequest{
//		TTSListID: req.TTSListID,
//		Numbers:       req.Numbers,
//		UserID:        req.UserID,
//	})
//	fmt.Println(err)
//	uj,_ := json.Marshal(result)
//
//	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Number purchased", "Number has been purchased",uj)
//	//return c.JSON(200, req)
//
//}
//
//func (uc *TelnyxNumbersGroupController) DeleteNumbers(c echo.Context) error {
//
//	// Stub an user to be populated from the body
//	var req ReqNumbers
//	if err := c.Bind(&req); err != nil {
//		return err
//	}
//	id := c.Param("id")
//	//phonenumbers := c.Param("phone_numbers")
//	err := uc.TelnyxService.DeleteNumbers(id)
//
//	fmt.Println(err)
//	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Number has been deleted", "Number found and deleted")
//	//return c.JSON(200, req)
//
//}
//
//func (uc *TelnyxNumbersGroupController) GetNumbers(c echo.Context) error {
//
//	cc := c.(model.AppContext)
//
//	// Grab user id for which file is being uploaded
//	page := c.QueryParam("page")
//	limit := c.QueryParam("limit")
//
//	fmt.Println(limit)
//	pageNumber, errPage := strconv.Atoi(page)
//	if errPage != nil  {
//		pageNumber = 1
//	}
//
//	limitNumber, errLimit := strconv.Atoi(limit)
//	if  errLimit != nil {
//		limitNumber = 0
//	}
//
//	fmt.Println(cc.ID)
//	fmt.Println(cc.ID.Hex())
//	fmt.Println(limitNumber)
//	campaignList, count, err := uc.TelnyxService.ListTelnyxNumbers(cc.ID.Hex(), pageNumber, limitNumber)
//	if err != nil {
//		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
//	}
//
//	pageInfo := model.Page{}
//	pageInfo.Page = pageNumber
//	pageInfo.Limit = limitNumber
//	pageInfo.TotalCount = count
//
//	// Marshal provided interface into JSON structure
//	data, _ := json.Marshal(campaignList)
//	pageData, _ := json.Marshal(pageInfo)
//
//	return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusOK, "Number List fetched successfully", "Number List fetched successfully", data, pageData)
//
//}
//
