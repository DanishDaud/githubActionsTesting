package apirouterv1

import (
	"github.com/go-redis/redis"
	"github.com/labstack/echo"
	"github.com/gomarkho/sas-rvm-provapi/controller/v1"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"gopkg.in/mgo.v2"
)

// this method takes restricted echo group
func RouteCampaignApis(acc *echo.Group, res *echo.Group, session *mgo.Session, redisSession *redis.Client, redisSessiontwo *redis.Client) {
	// initialize service of campaign
	campaignService := service.CampaignService{Service: service.Service{Session: session}, RedisService: service.RedisService{redisSessiontwo}}

	// initialize service of user
	userService := service.UserService{Service: service.Service{Session: session}}

	// initialize service of user detail
	userDetailService := service.UserDetailService{Service: service.Service{Session: session}}

	// initialize s3 service
	s3Service := service.NewS3Service()

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

	// initialize service of sound file
	soundFileService := service.SoundFileInfoService{Service: service.Service{Session: session}}

	ttsListService := service.TTSListService{Service: service.Service{Session: session}}
	cdrService := service.CdrService{Service: service.Service{Session: session}}
	eventService := service.EventsService{Service: service.Service{Session: session}}


	campaignController := controllerv1.CampaignController{CampaignService: &campaignService,
		UserService:         &userService,
		UserDetailService:   &userDetailService,
		SoundFileService:    &soundFileService,
		S3Service:           s3Service,
		ContactListService:  &contactListService,
		ContactGroupService: &contactGroupService,
		CDRService:          &cdrService,
		TTSListService:      &ttsListService,
		EventService: &eventService ,
	}

	// create campaign
	res.POST("/campaign", campaignController.CreateCampaign)

	// get campaign list
	res.GET("/campaign", campaignController.GetCampaignListForUser)

	// get campaign stat
	res.POST("/campaignstat", campaignController.GetCampaignStat)

	res.POST("/campaign/:cid/duplicate", campaignController.DuplicateCampaign)

	//	// update campaign name
	//	res.POST("/user/:id/campaign/:cid/name", campaignController.UpdateName)
	//

	// recycle campaign
	res.POST("/campaign/:cid/recycle", campaignController.Recyclecampaign)

	//	// export
	//	res.GET("/user/:id/campaign/:cid/export", campaignController.Export)
	//

	// update campaign name
	res.DELETE("/campaign/:cid", campaignController.DeleteCampaign)

	//	// rerun campaign
	//	res.POST("/user/:id/campaign/:cid/rerun", campaignController.RerunCampaign)
	//
	//	// update campaign
	res.PUT("/campaign/:cid", campaignController.UpdateCampaign)

	res.PUT("/campaign/:cid/speed", campaignController.UpdateCampaignSpeed)
	//
	//	// update call limits
	//	res.POST("/user/:id/campaign/:cid/limit", campaignController.UpdateLimits)

	// campaign start/stop
	res.GET("/campaign/:cid", campaignController.CampaignAction)

	// campaign detail
	res.GET("/campaign/:cid/detail", campaignController.CampaignDetail)
	//
	//	// test campaign by sending test call on given number
	res.POST("/campaign/:cid/test", campaignController.TestCampaign)

	res.POST("/campaign/:cid/test/tts", campaignController.TestTTSCampaign)

}
