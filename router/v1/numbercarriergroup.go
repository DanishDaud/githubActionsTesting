package apirouterv1

import (
	"github.com/go-redis/redis"
	"github.com/labstack/echo"
	controllerv1 "github.com/gomarkho/sas-rvm-provapi/controller/v1"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"gopkg.in/mgo.v2"
)

func RouteNumbersCarrierGroupApis(acc *echo.Group, res *echo.Group, session *mgo.Session, redisSession *redis.Client) {

	// initialize s3 service
	s3Service := service.NewS3Service()

	// initialize TCPA Service
	tcpaService := service.NewTCPAService()

	numbersCarrierGroupService := service.NumbersCarrierGroupService{
		Service: service.Service{Session: session},
		NumberService: service.NumberService{
			Service:     service.Service{Session: session},
			UserService: service.UserService{service.Service{Session: session}},
		},
		ContactListService: service.ContactListService{Service: service.Service{Session: session},
			RedisService: service.RedisService{redisSession},
			S3Service:    s3Service,
			TCPAService:  tcpaService,
		},
		ContactGroupService: service.ContactGroupService{Service: service.Service{Session: session}},
	}

	urc := controllerv1.NumbersCarrierGroupController{ //UserRoleService: &userRoleService,
		NumbersCarrierGroupService: &numbersCarrierGroupService,
	}

	res.POST("/numbercarriergroup/:id", urc.CreateSingle)
	res.POST("/numbercarriergroup", urc.CreateMultiple)
	//res.GET("/telnyxnumbersgroup", urc.GetSingle)
	res.GET("/numbercarriergroup", urc.GetAll)
	res.GET("/numbercarriergroup/contactlist", urc.GetContactListStats)
	res.DELETE("/numbercarriergroup/:id", urc.DeleteSingle)
	res.DELETE("/numbercarriergroup", urc.DeleteMultiple)

	//res.GET("/telnyx/searchnumbers", urc.SearchNumbers)
	//res.GET("/telnyxnumbers", urc.GetNumbers)
	//res.DELETE("/telnyxnumbers/:id", urc.DeleteNumbers)
}
