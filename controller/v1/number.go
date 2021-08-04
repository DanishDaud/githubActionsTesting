package controllerv1

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"github.com/gomarkho/sas-rvm-provapi/model"
	httpresponse "github.com/gomarkho/sas-rvm-provapi/network/response"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlmessages"
	"net/http"
	"strconv"
)

type NumberController struct {
	Controller
	//	UserRoleService      *service.UserRoleService
	NumberService      *service.NumberService
	TelnyxNumbersGroup *service.NumbersCarrierGroupService
	BillingManager     *service.BillingManager
}

func (uc *NumberController) SearchNumbers(c echo.Context) error {

	npa := c.QueryParam("npa")

	// so there would be no error in this case
	result, err := uc.NumberService.SearchBulkVSNumbers(model.BulkVsSearchRequest{
		Npa: npa,
	})
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}

	//uj,_ := json.Marshal(result)
	uj, _ := json.Marshal(result)
	//pageData, _ := json.Marshal(result)
	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Available numbers fetched successfully.", "Available numbers fetched successfully. From Carrier", uj)
}

//// CreateUser creates a  ̰new user resource
//func (uc *TelnyxController) SearchNumbers(c echo.Context) error {
//
//	// Stub an user to be populated from the body
//	countrycode := c.QueryParam("country_code")
//	locality := c.QueryParam("locality")
//	administrivearea := c.QueryParam("administrative_area")
//	number_type := c.QueryParam("number_type")
//	limit := c.QueryParam("limit")
//	npa := c.QueryParam("npa")
//
//	// so there would be no error in this case
//	result, err := uc.TelnyxService.SearchTelnyxNumbers(model.TelnyxSearchNumbersRequest{
//		NumberTypeStr: number_type,
//		AreaCode:         countrycode,
//		Locality:         locality,
//		AdministriveArea: administrivearea,
//		Limit:            limit,
//		Npa: 				npa,
//	})
//	if err != nil {
//		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
//	}
//
//	//uj,_ := json.Marshal(result)
//	uj,_ := json.Marshal(result.Data)
//	pageData, _ := json.Marshal(result.Metadata)
//	return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusOK, "Available numbers fetched successfully.", "Please resend account verification email", uj,pageData)
//}

func (uc *NumberController) PurchaseNumbers(c echo.Context) error {
	// Stub an user to be populated from the body
	var req model.PurchaseNumbersRequest
	cc := c.(model.AppContext)
	if err := c.Bind(&req); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	req.UserID = cc.ID
	//phonenumbers := c.Param("phone_numbers")
	result, amount, err := uc.NumberService.PurchaseBulkVSNumbers(model.BulkvsPurchaseNumbersRequest{
		//TTSListID: req.TTSListID,
		Numbers: req.Numbers,
		UserID:  req.UserID,
	})
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.CarrierError, err.Error())
	}

	uj, _ := json.Marshal(result)
	err1 := uc.BillingManager.ProcessDeductBalance(cc.ID.Hex(), amount)
	if err1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.CarrierError, err.Error())
	}
	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Numbers purchased successfully", "Number has been purchased", uj)
	//return c.JSON(200, req)

}

//func (uc *NumberController) GetForwardingNumber(c echo.Context) error {
//
//	// Stub an user to be populated from the body
//	var req model.PurchaseNumbersRequest
//	number := c.(model.AppContext)
//	if err := c.Bind(&req); err != nil {
//		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
//	}
//	req.UserID = cc.ID
//	//phonenumbers := c.Param("phone_numbers")
//	result,err := uc.NumberService.PurchaseBulkVSNumbers(model.BulkvsPurchaseNumbersRequest{
//		//TTSListID: req.TTSListID,
//		Numbers:       req.Numbers,
//		UserID:        req.UserID,
//	})
//	if err != nil {
//		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.CarrierError, err.Error())
//	}
//	fmt.Println(err)
//	uj,_ := json.Marshal(result)
//
//	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Numbers purchased successfully", "Number has been purchased",uj)
//	//return c.JSON(200, req)
//
//}

//func (uc *TelnyxController) PurchaseNumbers(c echo.Context) error {
//
//	// Stub an user to be populated from the body
//	var req model.PurchaseNumbersRequest
//	cc := c.(model.AppContext)
//	if err := c.Bind(&req); err != nil {
//		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
//	}
//	req.UserID = cc.ID
//	//phonenumbers := c.Param("phone_numbers")
//	result,err := uc.TelnyxService.PurchaseTelnyxNumbers(model.PurchaseNumbersRequest{
//		//TTSListID: req.TTSListID,
//		Numbers:       req.Numbers,
//		UserID:        req.UserID,
//	})
//	if err != nil {
//		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.CarrierError, err.Error())
//	}
//	fmt.Println(err)
//	uj,_ := json.Marshal(result)
//
//	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Numbers purchased successfully", "Number has been purchased",uj)
//	//return c.JSON(200, req)
//
//}

func (uc *NumberController) GetForwardingNumber(c echo.Context) error {

	logrus.Debugln("Service hit received")
	number := c.Param("id")

	forwardingnumber, err := uc.NumberService.GetForwardingNumber(number)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	data, _ := json.Marshal(forwardingnumber)

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Number List fetched successfully.", "Number List fetched successfully", data)

}

func (uc *NumberController) GetAllNumbers(c echo.Context) error {

	cc := c.(model.AppContext)

	// Grab user id for which file is being uploaded
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	search := c.QueryParam("search")

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

	campaignList, count, err := uc.NumberService.GetAll(model.GetAllRequest{
		UserID: cc.ID,
		Page:   pageNumber,
		Limit:  limitNumber,
		Search: search,
	})
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

	return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusOK, "Number List fetched successfully.", "Number List fetched successfully", data, pageData)

}

func (uc *NumberController) UpdateMultipleNumbers(c echo.Context) error {

	// Stub an user to be populated from the body
	var req model.UpdateMultipleNumberRequest
	cc := c.(model.AppContext)
	if err := c.Bind(&req); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	//phonenumbers := c.Param("phone_numbers")
	if len(req.ForwardingNumber) < 10 {
		err := errors.New("Enter Digit Again Again")
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseEnterDigitAgain, err.Error())
	}
	if len(req.ForwardingNumber) > 10 {
		s := len(req.ForwardingNumber) - 10
		req.ForwardingNumber = req.ForwardingNumber[s:]
	}
	result, err := uc.NumberService.UpdateMultipleNumbers(model.UpdateMultipleNumberRequest{
		//TTSListID: req.TTSListID,
		NumberID:         req.NumberID,
		ForwardingNumber: req.ForwardingNumber,
		UserID:           cc.ID,
	})
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	uj, _ := json.Marshal(result)
	//
	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Forwarding number added successfully.", "Number has been purchased", uj)
	//return c.JSON(200, req)

}

func (uc *NumberController) UpdateSingleNumber(c echo.Context) error {
	var req model.UpdateSingleNumberRequest
	if err := c.Bind(&req); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	if len(req.ForwardingNumber) < 10 {
		err := errors.New("Enter Digit Again Again")
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseEnterDigitAgain, err.Error())
	}
	id := c.Param("id")

	if len(req.ForwardingNumber) > 10 {
		s := len(req.ForwardingNumber) - 10
		req.ForwardingNumber = req.ForwardingNumber[s:]
	}
	result, err := uc.NumberService.UpdateSingleNumber(model.UpdateSingleNumberRequest{
		ForwardingNumber: req.ForwardingNumber,
		NumberID:         id,
	})
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	uj, _ := json.Marshal(result)
	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Forwarding number added successfully", "Number found and deleted", uj)
	//return c.JSON(200, req)

}

func (uc *NumberController) DeleteSingleNumber(c echo.Context) error {

	// Stub an user to be populated from the body
	var req model.DeleteNumbersRequest
	if err := c.Bind(&req); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.NumberNotFound, err.Error())
	}
	id := c.Param("id")
	//phonenumbers := c.Param("phone_numbers")
	err := uc.NumberService.DeleteBulkVSSingleNumber(id)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.NumberNotFound, err.Error())
	}
	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Number deleted successfully.", "Number found and deleted")
	//return c.JSON(200, req)

}

func (uc *NumberController) DeleteMultipleNumbers(c echo.Context) error {

	// Stub an user to be populated from the body
	var req model.DeleteNumbersRequest
	if err := c.Bind(&req); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.NumberNotFound, err.Error())
	}
	err := uc.NumberService.DeleteMultipleBulkVSNumbers(model.DeleteNumbersRequest{
		NumberID: req.NumberID,
		UserID:   req.UserID,
	})
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.NumberNotFound, err.Error())
	}
	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Number deleted successfully.", "Number found and deleted")
	//return c.JSON(200, req)

}
