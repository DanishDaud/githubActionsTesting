package apirouterv1

import (
	"gopkg.in/mgo.v2"

	"github.com/labstack/echo"
)

func RouteUserDetailApis(acc *echo.Group, res *echo.Group, session *mgo.Session) {

	// initialize service of user role
	//	userRoleService := service.UserRoleService{Service: service.Service{Session: session}}

	// initialize service of user
	//userService := service.UserService{Service: service.Service{Session: session}}

	// initialize service of user detail
	//userDetailService := service.UserDetailService{Service: service.Service{Session: session}}

	// initialize user controller
	//	urc := controllerv1.UserDetailController{UserRoleService: &userRoleService, UserService: &userService, UserDetailService: &userDetailService}

	// get user detail
	//res.GET("/user/:id/userDetail", urc.GetUserDetailInfo)
}
