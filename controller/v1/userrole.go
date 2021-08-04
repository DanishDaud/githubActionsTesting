package controllerv1

import (
	"github.com/gomarkho/sas-rvm-provapi/data/service"
)

type UserRoleController struct {
	Controller
	//	UserRoleService *service.UserRoleService
	UserService *service.UserService
}
