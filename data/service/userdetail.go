package service

import (
	"errors"

	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"gopkg.in/mgo.v2/bson"
)

type UserDetailService struct {
	Service
}

// returns user database object of present in database
// otherwise it returns error
func (us *UserDetailService) UserDetailObject(objectID string) (*model.UserDetail, error) {
	// get new instance of user data source
	uds := us.userDetailDataSource()
	defer uds.Session.Close()

	if !bson.IsObjectIdHex(objectID) {
		return nil, errors.New("object id is not valid")
	}

	userObjectId := bson.ObjectIdHex(objectID)

	// get user object from database
	return uds.UserDetailObject(userObjectId)
}

// returns user detail database object of present in database
// otherwise it returns error
func (us *UserDetailService) UserDetailObjectWithUser(objectID string) (*model.UserDetail, error) {
	// get new instance of user data source
	uds := us.userDetailDataSource()
	defer uds.Session.Close()

	if !bson.IsObjectIdHex(objectID) {
		return nil, errors.New("object id is not valid")
	}

	userObjectId := bson.ObjectIdHex(objectID)

	// get user object from database
	return uds.UserDetailWithUser(userObjectId)
}

func (us *UserDetailService) SaveUserDetail(userDetail *model.UserDetail) error {
	// get new instance of user data source
	uds := us.userDetailDataSource()
	defer uds.Session.Close()

	return uds.SaveUserDetail(userDetail)
}

// return instance if user  data source
// every time a new instance would be created
func (us *UserDetailService) userDetailDataSource() *datasource.UserDetailDataSource {
	return &datasource.UserDetailDataSource{DataSource: datasource.DataSource{Session: us.Session.Copy()}}
}
