package apirouterv1

import (
	"github.com/go-redis/redis"
	"github.com/labstack/echo"
	controllerv1 "github.com/gomarkho/sas-rvm-provapi/controller/v1"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"gopkg.in/mgo.v2"
)

func RouteNumberApis(acc *echo.Group, res *echo.Group, session *mgo.Session, redisSession *redis.Client) {

	bm := service.NewBillingManager()

	numberService := service.NumberService{
		Service:     service.Service{Session: session},
		UserService: service.UserService{service.Service{Session: session}},
	}

	urc := controllerv1.NumberController{ //UserRoleService: &userRoleService,
		NumberService:  &numberService,
		BillingManager: bm,
	}

	res.GET("/number/searchnumbers", urc.SearchNumbers)
	res.GET("/number", urc.GetAllNumbers)
	res.GET("/number/:id", urc.GetForwardingNumber)
	//res.GET("/number/:id", urc.GetForwardingNumber)
	res.POST("/number", urc.PurchaseNumbers)
	res.PATCH("/number/:id", urc.UpdateSingleNumber)
	res.PATCH("/number", urc.UpdateMultipleNumbers)
	res.DELETE("/number/:id", urc.DeleteSingleNumber)
	res.DELETE("/number", urc.DeleteMultipleNumbers)
}
