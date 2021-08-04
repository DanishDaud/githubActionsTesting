package datasource

import (
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const cDefaultSoundFileInfos = "defaultsounds"

// Data Structure to represent SoundFileDatasource
type DefaultSoundFileInfoDataSource struct {
	DataSource
}

func (dsfds *DefaultSoundFileInfoDataSource) SoundFiles() (*model.DefaultSoundFileInfos, error) {

	soundFileInfoObjects := model.DefaultSoundFileInfos{}
	query := dsfds.DbSession().DB(cmlutils.DefaultDatabase()).C(cDefaultSoundFileInfos).Find(bson.M{})
	if err := query.All(&soundFileInfoObjects); err != nil {
		return nil, err
	}

	return &soundFileInfoObjects, nil
}

func (dsfds *DefaultSoundFileInfoDataSource) SoundFile(stype int) (*model.DefaultSoundFileInfo, error) {

	soundFileInfoObject := model.DefaultSoundFileInfo{}
	query := dsfds.DbSession().DB(cmlutils.DefaultDatabase()).C(cDefaultSoundFileInfos).Find(bson.M{"type": stype})
	if err := query.One(&soundFileInfoObject); err != nil {
		return nil, err
	}

	return &soundFileInfoObject, nil
}
