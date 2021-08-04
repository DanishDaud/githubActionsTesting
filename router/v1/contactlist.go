package apirouterv1

import (
	"github.com/go-redis/redis"
	"github.com/labstack/echo"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	controllerv1 "github.com/gomarkho/sas-rvm-provapi/controller/v1"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"gopkg.in/mgo.v2"
)

func RouteContactListApis(acc *echo.Group, res *echo.Group, session *mgo.Session, redisSession *redis.Client) {

	// initialize s3 service
	s3Service := service.NewS3Service()

	fileService := service.FileService{}

	// initialize TCPA Service
	tcpaService := service.NewTCPAService()

	// initialize service of contact list
	contactListService := service.ContactListService{Service: service.Service{Session: session},
		RedisService: service.RedisService{redisSession},
		S3Service:    s3Service,
		TCPAService:  tcpaService,
	}

	// initialize service of contact group
	contactGroupService := service.ContactGroupService{Service: service.Service{Session: session}}

	// initialize service of user
	userService := service.UserService{Service: service.Service{Session: session}}

	// initialize service of user detail
	userDetailService := service.UserDetailService{Service: service.Service{Session: session}}

	// initialize dnc job service
	dncJobService := service.DNCJobService{Service: service.Service{Session: session}, TCPAService: tcpaService}

	// initialize dnc job result service
	dncJobResultService := service.DNCJobResultService{Service: service.Service{Session: session}, RedisService: service.RedisService{redisSession}}

	// initialize service of campaign
	campaignService := service.CampaignService{Service: service.Service{Session: session}}

	numbersCarrierGroupService := service.NumbersCarrierGroupService{
		Service: service.Service{Session: session},
		NumberService: service.NumberService{
			Service:     service.Service{Session: session},
			UserService: userService,
		},
		ContactListService:  contactListService,
		ContactGroupService: contactGroupService,
	}

	go func() {
		cr := cron.New()
		_, err := cr.AddFunc("@every 5s", func() {
			dncJobService.DNCScrubJobCron()
		})
		if err != nil {
			logrus.Errorln(err.Error())
			//panic(err)
		}
		cr.Run()
		defer cr.Stop()
	}()

	go func() {
		cr1 := cron.New()

		_, err := cr1.AddFunc("@every 5s", func() {
			dncJobResultService.DNCJobResultCron()
		})
		if err != nil {
			logrus.Errorln(err.Error())
		}
		cr1.Run()
		defer cr1.Stop()
	}()

	// initialize contact list controller
	contactListController := controllerv1.ContactListController{ContactListService: &contactListService,
		ContactGroupService:        &contactGroupService,
		UserService:                &userService,
		UserDetailService:          &userDetailService,
		S3Service:                  s3Service,
		TCPAService:                tcpaService,
		CampaignService:            &campaignService,
		NumbersCarrierGroupService: &numbersCarrierGroupService,
		DNCJobsService:             &dncJobService,
		DNCJobResultService:        &dncJobResultService,
		FileService:                &fileService,
	}

	// csv/xls/xlxs file upload
	res.POST("/contactlist/new", contactListController.CallerGroupFromPurchasedNumbers)

	res.POST("/contactlist", contactListController.UploadContactListForUser)

	// get contact list
	res.GET("/contactlist", contactListController.GetContactListsForUser)

	// delete contact list
	res.DELETE("/contactlist/:cid", contactListController.ContactListDelete)
	//
	// get contact list detail
	res.GET("/contactlist/:cid", contactListController.ContactListDetail)

	res.POST("/contactlist/:cid", contactListController.AddDncNumber)

	res.DELETE("/contactlist", contactListController.ContactListDeleteMultiple)
}
