package controllerv1

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"github.com/gomarkho/sas-rvm-provapi/network/response"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlmessages"
)

type UserDetailController struct {
	Controller
	//UserRoleService   *service.UserRoleService
	UserService       *service.UserService
	UserDetailService *service.UserDetailService
}

func (udc *UserDetailController) GetUserDetailInfo(c echo.Context) error {

	// Grab user id for which file is being uploaded
	userID := c.Param("id")

	// get user object
	user, er1 := udc.UserService.UserObject(userID)
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	userDetail, err := udc.UserDetailService.UserDetailObjectWithUser(user.ID.Hex())
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	// Marshal provided interface into JSON structure
	userData, _ := json.Marshal(userDetail)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
}
