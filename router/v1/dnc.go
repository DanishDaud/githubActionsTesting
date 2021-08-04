package apirouterv1

import (
	"os"

	"github.com/labstack/echo"
	"github.com/gomarkho/sas-rvm-provapi/controller/v1"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"gopkg.in/mgo.v2"
)

func RouteDncApis(acc *echo.Group, res *echo.Group, session *mgo.Session) {
	// initialize s3 service
	s3Service := service.NewS3Service()

	// initialize TCPA Service
	tcpaService := service.NewTCPAService()

	// initialize service of user
	userService := service.UserService{Service: service.Service{Session: session}}

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

	// initialize service of balance
	balanceService := service.BalanceService{Service: service.Service{Session: session}}

	// initialize service of payment
	paymentService := service.PaymentService{Service: service.Service{Session: session}}

	// initialize user controller
	dnc := controllerv1.UserController{ //UserRoleService: &userRoleService,
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
	}

	// search dnc number
	res.GET("/dnc", dnc.GetDNCNumber)
	res.DELETE("/contactlistid/:cid", dnc.DNCNumberDelete)

}
