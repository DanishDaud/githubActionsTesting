package controllerv1

import (
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"github.com/gomarkho/sas-rvm-provapi/network/response"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlmessages"
	"net/http"
	"strconv"
)

type SystemMessageController struct {
	UserService    *service.UserService
	MessageService *service.SystemMessageService
	Controller
}

// CreateUser creates a new user resource
func (smCon *SystemMessageController) GetMessage(c echo.Context) error {

	messagetype := c.Param("type")

	mtype, _ := strconv.Atoi(messagetype)
	systemmessage, err := smCon.MessageService.SystemMessageObject(mtype)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	data, _ := json.Marshal(systemmessage)
	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "success", "success", data)
}
