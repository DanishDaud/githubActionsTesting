package service

/*import (
	"gopkg.in/mgo.v2/bson"

	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
)

type UserRoleService struct {
	Service
}

func (urs *UserRoleService) GetRoleByType(roleType int8) (*model.UserRole, error) {
	// get new instance of user datasource
	userRoleDataSource := urs.userRoleDataSource()
	defer userRoleDataSource.Session.Close()

	return userRoleDataSource.GetRoleByType(roleType)
}

func (urs *UserRoleService) RoleObjectWithId(roleId bson.ObjectId) (*model.UserRole, error) {
	// get new instance of user datasource
	userRoleDataSource := urs.userRoleDataSource()
	defer userRoleDataSource.Session.Close()

	return userRoleDataSource.RoleWithId(roleId)
}

// return instance if user role datasource
// every time a new instance would be created
func (urs *UserRoleService) userRoleDataSource() *datasource.UserRoleDataSource {
	return &datasource.UserRoleDataSource{DataSource: datasource.DataSource{Session: urs.Session.Copy()}}
}

// return instance if user role datasource
// every time a new instance would be created
func (urs *UserRoleService) userDataSource() *datasource.UserDataSource {
	return &datasource.UserDataSource{DataSource: datasource.DataSource{Session: urs.Session.Copy()}}
}*/
