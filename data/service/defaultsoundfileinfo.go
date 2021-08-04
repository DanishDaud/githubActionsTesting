package service

import (
	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
)

// Data Structure to represent sound file info service
type DefaultSoundFileInfoService struct {
	Service
}

func (dsfs *DefaultSoundFileInfoService) SoundFiles() (*model.DefaultSoundFileInfos, error) {

	// get new instance of sound file datasource
	soundFileDataSource := dsfs.defaultSoundFileDatasource()
	defer soundFileDataSource.Session.Close()

	return soundFileDataSource.SoundFiles()
}

func (dsfs *DefaultSoundFileInfoService) SoundFile(stype int) (*model.DefaultSoundFileInfo, error) {

	// get new instance of sound file datasource
	soundFileDataSource := dsfs.defaultSoundFileDatasource()
	defer soundFileDataSource.Session.Close()

	return soundFileDataSource.SoundFile(stype)
}

// return instance of default sound file datasource
// every time a new instance would be created
func (dsfs *DefaultSoundFileInfoService) defaultSoundFileDatasource() *datasource.DefaultSoundFileInfoDataSource {
	return &datasource.DefaultSoundFileInfoDataSource{DataSource: datasource.DataSource{Session: dsfs.Session.Copy()}}
}
