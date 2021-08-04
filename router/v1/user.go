package apirouterv1

import (
	"github.com/go-redis/redis"
	"os"

	"github.com/labstack/echo"
	"github.com/gomarkho/sas-rvm-provapi/controller/v1"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"gopkg.in/mgo.v2"
)

func RouteUserApis(acc *echo.Group, res *echo.Group, session *mgo.Session, redisSession *redis.Client, redisSessiontwo *redis.Client) {

	// initialize service of user role
	//userRoleService := service.UserRoleService{Service: service.Service{Session: session}}

	// initialize service of user
	userService := service.UserService{Service: service.Service{Session: session}}

	// cdr service
	cdrService := service.CdrService{Service: service.Service{Session: session}}

	// initialize s3 service
	s3Service := service.NewS3Service()

	// initialize TCPA Service
	tcpaService := service.NewTCPAService()

	// initialize service of contact list
	contactlistService := service.ContactListService{Service: service.Service{Session: session},
		S3Service:   s3Service,
		TCPAService: tcpaService,
	}

	// initialize wepay checkout service
	wepayCheckoutService := service.WepayCheckoutService{Service: service.Service{Session: session}}

	// initialize file service
	fileService := service.NewFileService()

	// initialize service of email
	emailService := service.NewEmailService(os.Getenv("SEND_GRID_API"))

	// initialize service of token
	tokenService := service.NewTokenService()

	// billing manager
	bm := service.NewBillingManager()

	RedisService := service.RedisService{redisSessiontwo}
	// initialize service of balance
	balanceService := service.BalanceService{Service: service.Service{Session: session}}

	// initialize service of payment
	paymentService := service.PaymentService{Service: service.Service{Session: session}}
	eventService := service.EventsService{Service: service.Service{Session: session}}

	// initialize user controller
	urc := controllerv1.UserController{ //UserRoleService: &userRoleService,
		UserService:          &userService,
		ContactListService:   &contactlistService,
		WePayCheckoutService: &wepayCheckoutService,
		S3Service:            s3Service,
		FileService:          fileService,
		EmailService:         emailService,
		TokenService:         tokenService,
		BillingManager:       bm,
		BalanceService:       &balanceService,
		PaymentService:       &paymentService,
		CdrService:           &cdrService,
		RedisService:         &RedisService,
		EventService: &eventService,

	}

	// login
	acc.POST("/login", urc.CustomerLoginInfo)

	//// Create a new user
	acc.POST("/signup", urc.CreateUser)
	//
	//// GMail login
	//acc.POST("/gmailSignIn", urc.SignInWithGMailAccount)
	//
	//// get user with domain
	acc.GET("/domain", urc.GetUser)
	//
	//// login customer which falls under a customer
	//acc.POST("/user/:id/customerLogin", urc.CustomerLoginInfo)
	//
	//// login customer with his telephonic credentials
	//acc.POST("/user/teleLogin", urc.TelephonicLogin)
	//
	// get user info
	res.GET("/user", urc.GetUserInfo)

	// get user stats info
	res.GET("/user/stats", urc.GetUserStatsInfo)

	// get user call reports stats info
	res.GET("/user/callreports", urc.GetUserCallReports)

	// get user call dispositons stats info
	res.GET("/user/calldispositions", urc.GetUserCallDispositions)

	// get user call ASR&ALOC stats info
	res.GET("/user/callloc", urc.GetUserCallLoc)

	// get user call CallMin&CallCount stats info
	res.GET("/user/callcount", urc.GetUserCallCount)

	// get user Costs stats info
	res.GET("/user/callcosts", urc.GetUserCallCosts)

	//get cost
	res.GET("/user/cost", urc.Cost)

	// get user Aleg Call stats info
	res.GET("/user/Aleg", urc.GetUserAlegCall)

	// get user Aleg Call stats info
	res.GET("/user/Bleg", urc.GetUserBlegCall)

	//// update user info
	res.PATCH("/user/timezone", urc.UpdateTimeZone)

	//// update user password
	res.PATCH("/user", urc.UpdateUserAccount)
	//
	//// delete number from dnc
	//res.DELETE("/user/:id/dnc", urc.DeleteNumberFromDNC)
	//
	//// add number in dnc
	//res.POST("/user/:id/adddnc", urc.AddDncNumber)
	//
	//// add number in dnc bulk
	//res.POST("/user/:id/adddncbulk", urc.AddDncBulk)
	//
	//// get number from dnc
	//res.GET("/user/:id/dnc", urc.FindNumberFromDNC)
	//
	//// get current funds
	res.GET("/user/currentfunds", urc.CurrentFunds)

	//// get billing details
	//res.GET("/user/:id/billingdetails", urc.BillingDetails)
	//
	//// get billing parent
	//res.GET("/user/:id/billingParent", urc.BillingParent)
	//
	//// upload logo
	//res.POST("/user/:id/logo", urc.UploadLogo)
	//
	//// upload logo
	//res.DELETE("/user/:id/logo", urc.DeleteLogo)
	//
	//// send support message
	//res.POST("/user/:id/sendsupportmessage", urc.SendSupportMessage)
	//
	//// add time restrictions
	//res.POST("/user/:id/timerestrictions", urc.AddTimeRestrictions)
	//
	//// add time restrictions
	//res.POST("/user/:id/apitoken", urc.GenerateApiToken)
	//
	//// api token
	//res.GET("/user/:id/apitoken", urc.ApiToken)
	//
	//// we pay checkout
	//res.POST("/user/:id/wepay/checkout", urc.WePayCheckout)
	//
	//// stripe key
	//res.POST("/user/:id/strip/key", urc.StripeKey)
	//
	//// stripe checkout
	//acc.POST("/user/:id/strip/charge", urc.StripeCharge)
}
