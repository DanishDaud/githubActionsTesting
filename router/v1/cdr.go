package apirouterv1

import (
	"github.com/labstack/echo"
	controllerv1 "github.com/gomarkho/sas-rvm-provapi/controller/v1"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"gopkg.in/mgo.v2"
)

func RouteCdrApis(acc *echo.Group, res *echo.Group, session *mgo.Session) {
	// initialize service of campaign
	cdrService := service.CdrService{Service: service.Service{Session: session}}

	// initialize service of user
	userService := service.UserService{Service: service.Service{Session: session}}

	// initialize service of campaign
	campaignService := service.CampaignService{Service: service.Service{Session: session}}

	//initialize controller
	cdrController := controllerv1.CdrController{UserService: &userService, CDRService: &cdrService, CampaignService: &campaignService}

	// get cdr list
	res.GET("/cdr", cdrController.CdrList)

	//get daily stats
	res.GET("/cdr/stats/daily", cdrController.CdrDaily)

	//get hourly aleg stats
	res.GET("/cdr/stats/hourly/aleg", cdrController.CdrHourlyAleg)

	//get hourly bleg stats
	res.GET("/cdr/stats/hourly/bleg", cdrController.CdrHourlyBleg)

	//get result of comparecampaign
	res.POST("/cdr/compare/campaign", cdrController.CompareCampaign)

	//get campaign hourly stats
	res.GET("/cdr/stats/campaign/hourly", cdrController.CampaignHourly)
	//get campaign daily stats

	res.GET("/cdr/stats/campaign/daily", cdrController.CampaignDaily)

	res.POST("/cdr/export", cdrController.ExportCdr)
}
