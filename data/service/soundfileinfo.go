package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

// Data Structure to represent sound file info service
type SoundFileInfoService struct {
	Service
}

func (sfs *SoundFileInfoService) SoundFileObjectWithId(objectID string) (*model.SoundFileInfo, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, errors.New("object id is not valid")
	}

	soundFileObjectId := bson.ObjectIdHex(objectID)

	// get new instance of sound file datasource
	soundFileDataSource := sfs.soundFileDatasource()
	defer soundFileDataSource.Session.Close()

	return soundFileDataSource.SoundFileInfoWithId(soundFileObjectId)
}

func (sfs *SoundFileInfoService) SoundFileListWithUserId(objectID string, page int, limit int) (*model.SoundFileInfos, int, error) {
	if !bson.IsObjectIdHex(objectID) {
		return nil, 0, errors.New("Object id is not valid")
	}

	userObjectId := bson.ObjectIdHex(objectID)

	// get new instance of sound file datasource
	soundFileDataSource := sfs.soundFileDatasource()
	defer soundFileDataSource.Session.Close()

	return soundFileDataSource.SoundFileGetList(userObjectId, page, limit)
}

func (sfs *SoundFileInfoService) SaveSoundFile(soundFileInfo *model.SoundFileInfo) error {
	// get new instance of sound file datasource
	soundFileDataSource := sfs.soundFileDatasource()
	defer soundFileDataSource.Session.Close()

	return soundFileDataSource.SaveSoundFile(soundFileInfo)
}

func (sfs *SoundFileInfoService) DeleteSoundFile(soundFileInfo *model.SoundFileInfo) error {
	// get new instance of sound file datasource
	soundFileDataSource := sfs.soundFileDatasource()
	defer soundFileDataSource.Session.Close()

	return soundFileDataSource.DeleteSoundFile(soundFileInfo)
}

// this method takes the input file and generated pcmu, pcma, and .wav files
func (sfs *SoundFileInfoService) PerformSoundFileConversions(sourcePath string) (error, string, string) {
	fileBaseName := filepath.Base(sourcePath)
	fileName := strings.TrimSuffix(fileBaseName, filepath.Ext(fileBaseName))

	// create new destination path
	destinationPath := "/tmp/"

	wavDestinationPath := destinationPath + fileName + ".wav"
	pcmaDestinationPath := destinationPath + fileName + ".PCMU"

	logrus.Infoln(".wav destination path : " + wavDestinationPath + "\n")

	// generate .wav file
	command := fmt.Sprintf("ffmpeg -i %s -acodec pcm_s16le -ac 1 -ar 8000 %s >/dev/null 2>&1", sourcePath, wavDestinationPath)

	logrus.Infoln(".wav conversion command : " + command)

	out, error1 := exec.Command("sh", "-c", command).Output()
	if error1 != nil {
		logrus.Errorln("Error converting file to .wav : ")
		logrus.Errorln(error1.Error())
		logrus.Infoln("\n")
		return errors.New("Error converting file to .wav"), "", ""
	}

	fmt.Printf("%s\n\n", out)

	// generate .pcma file
	fsPCMA := fmt.Sprintf("sox %s -t raw -r 8k -b 8 -c 1 -e u-law %s", wavDestinationPath, pcmaDestinationPath)
	logrus.Infoln("pcma command : " + fsPCMA + "\n")
	fsPCMAO, error2 := exec.Command("sh", "-c", fsPCMA).Output()
	if error2 != nil {
		cmlutils.DeleteFile(wavDestinationPath)
		logrus.Errorln("Error converting file to .pcma")
		logrus.Errorln(error2.Error())
		return errors.New("Error converting file to .PCMA"), "", ""
	}
	fmt.Printf("%s\n\n", fsPCMAO)

	return nil, wavDestinationPath, pcmaDestinationPath
}

// this method checks if input file extension is valid or not
// allowed extensions are .mp3 and .wav
func (sfs *SoundFileInfoService) IsValidSoundFileFormat(fileName string) bool {
	isValidFileFormat := false
	fname := strings.ToLower(fileName)

	if strings.HasSuffix(fname, ".mp3") || strings.HasSuffix(fname, ".wav") {
		isValidFileFormat = true
	}

	return isValidFileFormat
}

// return instance of sound file datasource
// every time a new instance would be created
func (sfs *SoundFileInfoService) soundFileDatasource() *datasource.SoundFileInfoDataSource {
	return &datasource.SoundFileInfoDataSource{DataSource: datasource.DataSource{Session: sfs.Session.Copy()}}
}
