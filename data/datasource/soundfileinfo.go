package datasource

import (
	"errors"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const cSoundFileInfos = "soundfileinfos"

// Data Structure to represent SoundFileDatasource
type SoundFileInfoDataSource struct {
	DataSource
}

func (sfds *SoundFileInfoDataSource) SoundFileInfoWithId(objectId bson.ObjectId) (*model.SoundFileInfo, error) {
	soundFileInfo := model.SoundFileInfo{}
	err := sfds.DbSession().DB(cmlutils.DefaultDatabase()).C(cSoundFileInfos).FindId(objectId).One(&soundFileInfo)
	if err != nil {
		return nil, errors.New("Can not find soundfile")
	}

	return &soundFileInfo, nil
}

func (sfds *SoundFileInfoDataSource) SoundFileInfoWithIdAndUserId(userId bson.ObjectId, soundObjectId bson.ObjectId) (*model.SoundFileInfo, error) {
	soundFileInfo := model.SoundFileInfo{}
	err := sfds.DbSession().DB(cmlutils.DefaultDatabase()).C(cSoundFileInfos).Find(bson.M{"_id": soundObjectId, "userid": userId}).One(&soundFileInfo)
	if err != nil {
		logrus.Errorln(err.Error())
		return nil, errors.New("Can not find soundfile")
	}

	return &soundFileInfo, nil
}

func (sfds *SoundFileInfoDataSource) SoundFileGetList(userObjectId bson.ObjectId, page int, limit int) (*model.SoundFileInfos, int, error) {

	soundFileInfoObjects := model.SoundFileInfos{}

	query := sfds.DbSession().DB(cmlutils.DefaultDatabase()).C(cSoundFileInfos).Find(bson.M{"userid": userObjectId})
	query1 := sfds.DbSession().DB(cmlutils.DefaultDatabase()).C(cSoundFileInfos).Find(bson.M{"userid": userObjectId})

	//if page < 1 || limit < 1 {
	//	if err := query.All(&soundFileInfoObjects); err != nil {
	//		return nil, 0, err
	//	}
	//} else {
	if err := query.Sort("-createDate").Skip((page - 1) * limit).Limit(limit).All(&soundFileInfoObjects); err != nil {
		return nil, 0, err
	}
	//}

	count, er := query1.Count()
	if er != nil {
		return nil, 0, er
	}

	return &soundFileInfoObjects, count, nil
}

func (sfds *SoundFileInfoDataSource) SaveSoundFile(soundFile *model.SoundFileInfo) error {
	// if there is no user id assign one

	if soundFile.ID == "" {
		soundFile.ID = bson.NewObjectId()
	}

	existingSoundFileObject, err := sfds.SoundFileInfoWithId(soundFile.ID)
	if err != nil {
		// user doesn't exist, create new
		soundFile.CreateDate = time.Now().UTC()
		soundFile.UpdateDate = time.Now().UTC()
		// Write the user to mongo
		if err := sfds.DbSession().DB(cmlutils.DefaultDatabase()).C(cSoundFileInfos).Insert(soundFile); err != nil {
			return err
		}
	} else {
		// user exists
		soundFile.UpdateDate = time.Now().UTC()
		if err := sfds.DbSession().DB(cmlutils.DefaultDatabase()).C(cSoundFileInfos).UpdateId(existingSoundFileObject.ID, soundFile); err != nil {
			return err
		}
	}
	return nil
}

func (sfds *SoundFileInfoDataSource) DeleteSoundFile(soundFileObject *model.SoundFileInfo) error {
	// delete sound file object from database
	if err := sfds.DbSession().DB(cmlutils.DefaultDatabase()).C(cSoundFileInfos).RemoveId(soundFileObject.ID); err != nil {
		// sound object delete fail
		return err
	}
	return nil
}
