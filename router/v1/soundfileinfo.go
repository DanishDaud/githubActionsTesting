package apirouterv1

import (
	"github.com/labstack/echo"
	"github.com/gomarkho/sas-rvm-provapi/controller/v1"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"gopkg.in/mgo.v2"
)

func RouteSoundFileInfoApis(acc *echo.Group, res *echo.Group, session *mgo.Session) {

	// initialize service of sound file
	soundFileService := service.SoundFileInfoService{Service: service.Service{Session: session}}

	// initialize service of default sound file
	defaultSoundFileService := service.DefaultSoundFileInfoService{Service: service.Service{Session: session}}

	// initialize service of user
	userService := service.UserService{Service: service.Service{Session: session}}

	// initialize s3 service
	s3Service := service.NewS3Service()

	// initialize service of campaign
	campaignService := service.CampaignService{Service: service.Service{Session: session}}

	//stats := campaignService.CampaignLiveStats("5cc8307790e8803f00f7f9b6", "5da4ae0095a1be2361b8908b")
	//logrus.Infoln(stats)

	// initialize tts service
	ttsService := service.NewTTSService()

	// initialize user controller
	soundFileController := controllerv1.SoundFileInfoController{SoundFileInfoService: &soundFileService,
		DefaultSoundFileInfoService: &defaultSoundFileService,
		UserService:                 &userService,
		S3Service:                   s3Service,
		TTSService:                  ttsService,
		CampaignService:             &campaignService,
	}

	//// upload sound file
	res.POST("/soundfile", soundFileController.UploadSoundFile)
	//
	//// update sound file name for resource
	//res.PATCH("/user/:id/soundfile/:sid", soundFileController.SoundFileNameUpdate)
	//
	//// get sound file info
	res.GET("/soundfile/:sid", soundFileController.GetInfo)

	//
	//// delete sound file
	res.DELETE("/soundfile/:sid", soundFileController.SoundFileDelete)

	res.DELETE("/soundfile", soundFileController.SoundFileDeleteMultiple)
	//
	//// get sound file list
	res.GET("/soundfile", soundFileController.SoundFilesGetList)
	//
	//// get default sound file list
	//res.GET("/user/:id/soundfile/default", soundFileController.DefaultSoundFiles)
	//
	//// text to speech
	//res.POST("/user/:id/soundfile/tts", soundFileController.TextToSpeech)
}
