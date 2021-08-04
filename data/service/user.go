package service

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

type UserService struct {
	Service
}

// returns user database object if present in database
// otherwise it returns error
func (us *UserService) UserObjectWithCredentials(username string, password string) (*model.User, error) {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	// get user object from database
	return uds.UserObjectWithCredentials(username, password)
}

// returns user database object if present in database
// otherwise it returns error
func (us *UserService) UserObjectWithEmailAndBillingParent(email string, isVerified bool) (*model.User, error) {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	// get user object from database
	return uds.UserObjectEmailAndBillingParent(email, isVerified)
}

// change password of user with id
func (us *UserService) ChangePassword(userId string, newPassword string) error {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	if !bson.IsObjectIdHex(userId) {
		return errors.New("object id is not valid")
	}

	objId := bson.ObjectIdHex(userId)

	return uds.ChangePassword(objId, newPassword)
}

// change password of user with id
func (us *UserService) UpdateStats(userId string, stats model.Stats) error {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	if !bson.IsObjectIdHex(userId) {
		return errors.New("object id is not valid")
	}

	objId := bson.ObjectIdHex(userId)

	return uds.UpdateStats(objId, stats)
}

// change logo of user with id
func (us *UserService) UpdateLogo(userId string, logoPath string) error {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	if !bson.IsObjectIdHex(userId) {
		return errors.New("object id is not valid")
	}

	objId := bson.ObjectIdHex(userId)

	return uds.UpdateLogo(objId, logoPath)
}

// change logo of user with id
func (us *UserService) UpdateTimeRestrictions(userId string, tr model.Restrictions) error {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	if !bson.IsObjectIdHex(userId) {
		return errors.New("object id is not valid")
	}

	objId := bson.ObjectIdHex(userId)

	return uds.UpdateTimeRestrictions(objId, tr)
}

// returns user database object if present in database
// otherwise it returns error
func (us *UserService) VerifyUser(objectId string) error {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	if !bson.IsObjectIdHex(objectId) {
		return errors.New("object id is not valid")
	}

	objId := bson.ObjectIdHex(objectId)

	return uds.VerifyUser(objId)
}

// returns user database object if present in database
// otherwise it returns error
func (us *UserService) UserObjectWithGMailAccount(username string, gmailId string) (*model.User, error) {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	// get user object from database
	return uds.UserObjectWithGMailAccount(username, gmailId)
}

// returns user database object if present in database
// otherwise it returns error
func (us *UserService) UserObjectWithTelephonicCredentials(telephonicId int32, telephonicCode int32) (*model.User, error) {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	// get user object from database
	return uds.UserObjectWithTelephonicCredentials(telephonicId, telephonicCode)
}

// returns user database object of present in database
// otherwise it returns error
func (us *UserService) UserObject(objectID string) (*model.User, error) {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	if !bson.IsObjectIdHex(objectID) {
		return nil, errors.New("object id is not valid")
	}

	userObjectId := bson.ObjectIdHex(objectID)

	// get user object from database
	return uds.UserObject(userObjectId)
}

func (us *UserService) ActiveCampaign(objectID string) (int, error) {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	if !bson.IsObjectIdHex(objectID) {
		return -1, errors.New("object id is not valid")
	}

	userObjectId := bson.ObjectIdHex(objectID)

	// get user object from database
	return uds.ActiveCampaign(userObjectId)
}

func (us *UserService) TotalCost(objectID string) (float32, error) {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	if !bson.IsObjectIdHex(objectID) {
		return -1, errors.New("object id is not valid")
	}

	userObjectId := bson.ObjectIdHex(objectID)

	// get user object from database
	return uds.TotalCost(userObjectId)
}

func (us *UserService) UserDetailObjectWithUser(objectID string) (*model.UserDetail, error) {
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

// this method checks if user with the given email exists or not
func (us *UserService) IsUserExistsWithEmailAddress(email string) bool {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	_, err := uds.UserWithEmailAddress(email)
	if err != nil {
		return false
	}

	return true
}

func (us *UserService) UserWithEmailAddress(email string) (*model.User, error) {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	user, err := uds.UserWithEmailAddress(email)
	return user, err
}
func (us *UserService) UserWithType(usertype int) (*model.User, error) {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	user, err := uds.UserWithType(usertype)
	return user, err
}

func (us *UserService) UserByAPIToken(token string) (*model.User, error) {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	user, err := uds.UserByAPIToken(token)
	return user, err
}

// this method create new user in the database
// it doesn't check internally whether all the validation are applied or not
func (us *UserService) CreateNewUser(user *model.User) error {

	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	udds := us.userDetailDataSource()
	defer uds.Session.Close()

	// get new instance of contact list data source
	clds := us.contactListDataSource()
	defer uds.Session.Close()

	// encrypt password to .md5
	user.Password = cmlutils.GetMD5Hash(user.Password)

	if err := us.SaveUser(user); err != nil {
		return err
	}

	// for every user create a do not contact list for that user (only if user is customer)
	//if user.Role.Type == model.Client || user.Role.Type == model.Admin {
	// if new user is customer, create a do not contact list as well

	contactList := model.ContactList{}
	contactList.ID = bson.NewObjectId()
	contactList.FileName = "DNCL"
	contactList.Name = "DNCL"
	contactList.NumberCount = 0
	contactList.FileS3Path = ""
	contactList.UserID = user.ID
	//contactList.IsDoNotContactList = true

	// insert data into db
	err := clds.SaveContactList(&contactList)
	if err != nil {
		logrus.Errorln("dncl save failed for user : ", user.Email)
		logrus.Infoln("******* handle this error on priority, if contact list save fail remove user as well and return fail *********")
	}

	userDetail := model.UserDetail{UserID: user.ID}

	// save user detail
	udds.SaveUserDetail(&userDetail)
	//}

	return nil
}
func (us *UserService) SaveUserDetail(userDetail *model.UserDetail) error {
	// get new instance of user data source
	uds := us.userDetailDataSource()
	defer uds.Session.Close()

	return uds.SaveUserDetail(userDetail)
}
func (us *UserService) SaveUser(user *model.User) error {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	return uds.SaveUser(user)
}

// return owner of the given domain
func (us *UserService) GetDomainOwnerForUser(domain string) (*model.User, error) {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	return uds.UserWithDomain(domain)
}

func (us *UserService) UpdateNotificationPreference(objectId string, notfid string) error {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	if !bson.IsObjectIdHex(objectId) {
		return errors.New("object id is not valid")
	}

	userObjectId := bson.ObjectIdHex(objectId)

	return uds.UpdateNotificationPreference(userObjectId, notfid)
}

func (us *UserService) GenerateAPIToken(objectId string) (string, error) {
	// get new instance of user data source
	uds := us.userDataSource()
	defer uds.Session.Close()

	if !bson.IsObjectIdHex(objectId) {
		return "", errors.New("object id is not valid")
	}

	userObjectId := bson.ObjectIdHex(objectId)

	// get user object from database
	usr, err := uds.UserObject(userObjectId)
	if err != nil {
		return "", err
	}
	token := bson.NewObjectId().Hex() + "_" + usr.ID.Hex()
	err = uds.UpdateAPIToken(userObjectId, token)
	return token, err
}

// return instance if user  data source
// every time a new instance would be created
func (us *UserService) userDataSource() *datasource.UserDataSource {
	return &datasource.UserDataSource{DataSource: datasource.DataSource{Session: us.Session.Copy()}}
}

// return instance if user detail data source
// every time a new instance would be created
func (us *UserService) userDetailDataSource() *datasource.UserDetailDataSource {
	return &datasource.UserDetailDataSource{DataSource: datasource.DataSource{Session: us.Session.Copy()}}
}

// return instance if contact list  data source
// every time a new instance would be created
func (us *UserService) contactListDataSource() *datasource.ContactListDataSource {
	return &datasource.ContactListDataSource{DataSource: datasource.DataSource{Session: us.Session.Copy()}}
}
