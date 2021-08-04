package datasource

import (
	"time"

	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const cUserDetails = "userdetails"

type UserDetailDataSource struct {
	DataSource
}

// this method returns user database object using object id
func (uds *UserDetailDataSource) UserDetailObject(objectID bson.ObjectId) (*model.UserDetail, error) {
	parent := model.UserDetail{}

	err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUserDetails).FindId(objectID).One(&parent)

	if err != nil {
		return nil, err
	}
	return &parent, err
}

// this method returns user database object using object id
func (uds *UserDetailDataSource) UserDetailWithUser(userId bson.ObjectId) (*model.UserDetail, error) {
	user := model.UserDetail{}

	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUserDetails).
		Find(bson.M{"userid": userId}).One(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (uds *UserDetailDataSource) SaveUserDetail(userDetail *model.UserDetail) error {
	// if there is no user id assign one
	if userDetail.ID == "" {
		userDetail.ID = bson.NewObjectId()
	}

	existingUserDetailObject, err := uds.UserDetailObject(userDetail.ID)
	if err != nil {
		// user doesn't exist, create new
		userDetail.CreateDate = time.Now().UTC()
		userDetail.UpdateDate = time.Now().UTC()
		// Write the user to mongo
		if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUserDetails).Insert(userDetail); err != nil {
			return err
		}
	} else {
		// user exists
		userDetail.UpdateDate = time.Now().UTC()
		if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUserDetails).UpdateId(existingUserDetailObject.ID, userDetail); err != nil {
			return err
		}
	}
	return nil
}
