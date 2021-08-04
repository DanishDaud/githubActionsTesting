package datasource

import (
	"time"

	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const cUsers = "users"

type UserDataSource struct {
	DataSource
}

// this method returns user database object using object id
func (uds *UserDataSource) UserObject(objectID bson.ObjectId) (*model.User, error) {
	parent := model.User{}

	err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).FindId(objectID).One(&parent)

	if err != nil {
		return nil, err
	}
	return &parent, err
}

func (uds *UserDataSource) GetUserById(id bson.ObjectId) (*model.User, error) {
	user := model.User{}
	err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).Find(bson.M{"_id": id}).One(&user)
	if err != nil {
		return nil, err
	}
	return &user, err
}

// this method returns user database object using object id
func (uds *UserDataSource) UserObjectWithCredentials(username string, password string) (*model.User, error) {
	user := model.User{}

	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).
		Find(bson.M{"$and": []bson.M{
			bson.M{"email": username},
			bson.M{"password": cmlutils.GetMD5Hash(password)}}}).One(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (uds *UserDataSource) VerifyUser(userId bson.ObjectId) error {
	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).UpdateId(userId, bson.M{"$set": bson.M{"isVerified": true}}); err != nil {
		return err
	}
	return nil
}

func (uds *UserDataSource) ChangePassword(userId bson.ObjectId, newPassword string) error {
	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).UpdateId(userId, bson.M{"$set": bson.M{"password": cmlutils.GetMD5Hash(newPassword)}}); err != nil {
		return err
	}
	return nil
}

func (uds *UserDataSource) UpdateStats(userId bson.ObjectId, stats model.Stats) error {
	return uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).UpdateId(userId, bson.M{"$inc": bson.M{
		// call stats
		"userStats.dialed":           -stats.Dialed,
		"userStats.busy":             -stats.Busy,
		"userStats.busyDuration":     -stats.BusyDuration,
		"userStats.error":            -stats.Error,
		"userStats.errorDuration":    -stats.ErrorDuration,
		"userStats.noAnswer":         -stats.NoAnswer,
		"userStats.noAnsDuration":    -stats.NoAnsDuration,
		"userStats.live":             -stats.Live,
		"userStats.liveDuration":     -stats.LiveDuration,
		"userStats.voiceMail":        -stats.VoiceMail,
		"userStats.vmDuration":       -stats.VMDuration,
		"userStats.dnc":              -stats.DNC,
		"userStats.dncDuration":      -stats.DNCDuration,
		"userStats.transfer":         -stats.Transfer,
		"userStats.transferDuration": -stats.TransferDuration,
		"userStats.avgld":            -stats.Avgld,
	}})
}

// this method returns user database object using object id
func (uds *UserDataSource) UserObjectEmailAndBillingParent(email string, isVerified bool) (*model.User, error) {
	user := model.User{}

	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).
		Find(bson.M{"$and": []bson.M{
			{"email": email},
			{"isVerified": isVerified}}}).One(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// this method returns user database object using object id
func (uds *UserDataSource) UserObjectWithGMailAccount(username string, gmailId string) (*model.User, error) {
	user := model.User{}

	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).
		Find(bson.M{"$and": []bson.M{
			bson.M{"email": username},
			bson.M{"gmailId": gmailId}}}).One(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// this method returns user database object using object id
func (uds *UserDataSource) UserObjectWithTelephonicCredentials(telephonicId int32, telephonicCode int32) (*model.User, error) {
	user := model.User{}

	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).
		Find(bson.M{"$and": []bson.M{
			bson.M{"telephonicId": telephonicId},
			bson.M{"telephonicCode": telephonicCode}}}).One(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (uds *UserDataSource) UpdateLogo(userId bson.ObjectId, logoPath string) error {
	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).UpdateId(userId, bson.M{"$set": bson.M{"customize.logoPath": logoPath}}); err != nil {
		return err
	}
	return nil
}

func (uds *UserDataSource) UpdateTimeRestrictions(userId bson.ObjectId, tr model.Restrictions) error {
	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).UpdateId(userId, bson.M{"$set": bson.M{"restrictions": tr}}); err != nil {
		return err
	}
	return nil
}

func (uds *UserDataSource) SaveUser(user *model.User) error {
	// if there is no user id assign one
	if user.ID == "" {
		user.ID = bson.NewObjectId()
	}

	existingUserObject, err := uds.UserObject(user.ID)
	if err != nil {
		// user doesn't exist, create new
		user.CreateDate = time.Now().UTC()
		user.UpdateDate = time.Now().UTC()
		// Write the user to mongo
		if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).Insert(user); err != nil {
			return err
		}
	} else {
		// user exists
		user.UpdateDate = time.Now().UTC()
		if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).UpdateId(existingUserObject.ID, user); err != nil {
			return err
		}
	}
	return nil
}

// this method returns the user using his email address
func (uds *UserDataSource) UserWithEmailAddress(emailAddress string) (*model.User, error) {

	user := model.User{}

	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).Find(bson.M{"email": emailAddress}).One(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
func (uds *UserDataSource) UserWithType(userType int) (*model.User, error) {

	user := model.User{}

	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).Find(bson.M{"role": userType}).One(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// this method returns the user using his email address
func (uds *UserDataSource) UserByAPIToken(token string) (*model.User, error) {

	user := model.User{}

	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).Find(bson.M{"apiToken": token}).One(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// this method returns the owner of the domain
func (uds *UserDataSource) UserWithDomain(domain string) (*model.User, error) {
	user := model.User{}

	err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).Find(bson.M{"account.homePageUrl": domain}).One(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (uds *UserDataSource) UpdateNotificationPreference(objectId bson.ObjectId, notfid string) error {
	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).UpdateId(objectId, bson.M{"$set": bson.M{"payments.wePayInfo.notification_preference_id": notfid}}); err != nil {
		return err
	}
	return nil
}

func (uds *UserDataSource) UpdateAPIToken(objectId bson.ObjectId, token string) error {
	if err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cUsers).UpdateId(objectId, bson.M{"$set": bson.M{"apiToken": token}}); err != nil {
		return err
	}
	return nil
}
