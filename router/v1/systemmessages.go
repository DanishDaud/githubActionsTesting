package apirouterv1

import (
	"github.com/labstack/echo"
	"github.com/gomarkho/sas-rvm-provapi/controller/v1"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"gopkg.in/mgo.v2"
)

func RouteSystemMessagesApis(acc *echo.Group, res *echo.Group, session *mgo.Session) {

	userService := service.UserService{Service: service.Service{Session: session}}

	messageService := service.SystemMessageService{Service: service.Service{Session: session}}

	// instantiate new root controller
	smc := controllerv1.SystemMessageController{
		UserService:    &userService,
		MessageService: &messageService}

	res.GET("/systemmessage/:type", smc.GetMessage)

}
