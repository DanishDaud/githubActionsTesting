package apirouterv1

import (
	"os"

	"github.com/labstack/echo"
	"github.com/gomarkho/sas-rvm-provapi/controller/v1"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"gopkg.in/mgo.v2"
)

func RouteRootApis(acc *echo.Group, res *echo.Group, session *mgo.Session) {

	// initialize service of token
	tokenService := service.Token{}

	// initialize service of email
	emailService := service.NewEmailService(os.Getenv("SEND_GRID_API"))

	// initialize service of user
	userService := service.UserService{Service: service.Service{Session: session}}

	// instantiate new root controller
	rootController := controllerv1.RootController{TokenService: &tokenService,
		EmailService: emailService,
		UserService:  &userService}

	// get resources home address
	acc.GET("/assethome", rootController.GetAssetHomePath)
	// verify email
	//// verify client email without password
	acc.GET("/confirm", rootController.ConfirmEmail)
	//// confirm
	//// verify client email with password
	//// this is for verification of users created through admin panel
	acc.GET("/verify", rootController.VerifyEmailWithPassword)
	//// verify after reset
	acc.GET("/resetverify", rootController.VerifyEmailWithPassword)
	//// send forgot password email
	acc.GET("/sendforgot", rootController.SendForgotPasswordEmail)
	//// verify forgot password process
	acc.GET("/forgotpasswordverify", rootController.ForgotPasswordVerify)
	//// resent invite (for users with password)
	acc.GET("/resendverify", rootController.ResendVerify)
}
