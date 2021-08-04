package datasource

/*
import (

	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const cUserRoles = "userroles"

type UserRoleDataSource struct {
	DataSource
}

func (urDs *UserRoleDataSource) GetRoleByType(roleType int8) (*model.UserRole, error) {
	roleObject := model.UserRole{}

	// Fetch Role
	if err := urDs.DbSession().DB(cmlutils.DefaultDatabase()).C(cUserRoles).Find(bson.M{"type": roleType}).One(&roleObject); err != nil {
		return nil, err
	}
	return &roleObject, nil
}

// returns role with id
func (urDs *UserRoleDataSource) RoleWithId(roleId bson.ObjectId) (*model.UserRole, error) {
	roleObject := model.UserRole{}

	// Fetch Role
	if err := urDs.DbSession().DB(cmlutils.DefaultDatabase()).C(cUserRoles).FindId(roleId).One(&roleObject); err != nil {
		return nil, err
	}
	return &roleObject, nil
}*/
