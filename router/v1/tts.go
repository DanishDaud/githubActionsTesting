package apirouterv1

import (
	"github.com/go-redis/redis"
	"github.com/labstack/echo"
	"github.com/gomarkho/sas-rvm-provapi/controller/v1"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"gopkg.in/mgo.v2"
)

func RouteSystemTTSApis(acc *echo.Group, res *echo.Group, session *mgo.Session, redisSession *redis.Client) {

	s3Service := service.S3Service{}

	// initialize service of contact list
	contactListService := service.ContactListService{Service: service.Service{Session: session},
		RedisService: service.RedisService{RedisClient: redisSession},
	}
	CustomfieldService := service.CustomFieldService{Service: service.Service{Session: session}}

	TTSGroupService := service.TTSGroupService{Service: service.Service{
		Session: session,
	}}
	userService := service.UserService{Service: service.Service{
		Session: session,
	}}
	ttsService := service.TTSListService{Service: service.Service{Session: session},
		ContactListService: contactListService,
		CustomFieldService: CustomfieldService,
		TTSGroup:           TTSGroupService,
		UserService:        userService,
	}

	campaignService := service.CampaignService{Service: service.Service{
		Session: session,
	}}

	// instantiate new root controller
	smc := controllerv1.TTSController{
		TTSService:         &ttsService,
		S3Service:          &s3Service,
		ContactListService: &contactListService,
		UserService:        &userService,
		CampaignService:    &campaignService,
	}

	res.POST("/file", smc.SaveTTSFile)

	res.POST("/import", smc.Import)

	res.GET("/ttslist", smc.TTSList)

	res.DELETE("/ttslist/:tid", smc.DeleteTTS)

	res.GET("/tts/:tid", smc.DeatilTTSList)

	res.DELETE("/ttslist", smc.DeleteMultipleTTSList)
}
