package apirouterv1

import (
	"github.com/labstack/echo"
	controllerv1 "github.com/gomarkho/sas-rvm-provapi/controller/v1"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"gopkg.in/mgo.v2"
)

// this method takes restricted echo group
func RouteBillingApis(acc *echo.Group, res *echo.Group, session *mgo.Session) {
	// initialize service of user
	userService := service.PaymentService{Service: service.Service{Session: session}}
	urc := controllerv1.BillingController{ //UserRoleService: &userRoleService,
		PaymentService: &userService,
	}

	// get campaign list
	res.GET("/billing/history", urc.BillingHistory)
}
