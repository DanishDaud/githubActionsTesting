package apirouterv1

import (
	"github.com/labstack/echo"
	"github.com/gomarkho/sas-rvm-provapi/controller/v1"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"gopkg.in/mgo.v2"
)

func RouteCustomFieldApis(acc *echo.Group, res *echo.Group, session *mgo.Session) {

	customFieldService := service.CustomFieldService{Service: service.Service{Session: session}}

	// instantiate new root controller
	smc := controllerv1.CustomFieldController{
		CustomField: &customFieldService,
	}

	res.POST("/customfield", smc.Save)

	res.GET("/customfield", smc.List)
}
