package controllerv1

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/network/response"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlconstants"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlmessages"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

// Data Structure to represent Sound file controller
type SoundFileInfoController struct {
	Controller
	SoundFileInfoService        *service.SoundFileInfoService
	DefaultSoundFileInfoService *service.DefaultSoundFileInfoService
	UserService                 *service.UserService
	S3Service                   *service.S3Service
	TTSService                  *service.TTSService
	CampaignService             *service.CampaignService
}

type (
	NameUpdateObject struct {
		Name string `json:"name" validate:"required"`
	}

	TTSRequest struct {
		Text   string `json:"text" validate:"required"`
		Gender int    `json:"gender" validate:"required"`
	}
	SoundFileIDs struct {
		ID []string `json:"id"`
	}
)

// this method is for uploading sound file
// it will also upload file to s3
// name in the sound file object would same as uploaded file by default
func (sfc *SoundFileInfoController) UploadSoundFile(c echo.Context) error {
	cc := c.(model.AppContext)
	// Grab user id for which file is being uploaded
	name := cc.FormValue("name")
	userInfo, errUser := sfc.UserService.UserObject(cc.ID.Hex())
	if errUser != nil {
		return httpresponse.CreateBadResponse(&cc.Context, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	file, err := cc.FormFile("file")
	filename := file.Filename
	if err != nil {
		return httpresponse.CreateBadResponse(&cc.Context, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.SoundFileUploadFail)
	}

	// check file format
	if !sfc.SoundFileInfoService.IsValidSoundFileFormat(file.Filename) {
		return httpresponse.CreateBadResponse(&cc.Context, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.SoundFileFormatInCorrect)
	}

	// save incoming file
	destinationPath, errSave := sfc.saveMultipartFile(file)
	if errSave != nil {
		return httpresponse.CreateBadResponse(&cc.Context, http.StatusBadRequest, cmlmessages.OperationFailed, errSave.Error())
	}

	// get sound file duration
	duration, err := service.GetFileDuration(destinationPath)
	if err != nil {
		duration = 30
	}
	size, err := service.GetFileSize(destinationPath)
	if err != nil {
		size = 30
	}
	filetype := path.Ext(file.Filename)
	filetype = strings.Trim(filetype, ".")
	filetype = "audio/" + filetype
	// process sound file upload for user
	soundFileInfo, err1 := sfc.processSoundFile(filename, userInfo, name, destinationPath, duration, size, filetype)
	if err1 != nil {
		cmlutils.DeleteFile(destinationPath)
		return httpresponse.CreateBadResponse(&cc.Context, http.StatusBadRequest, cmlmessages.OperationFailed, err1.Error())
	}

	// Marshal provided interface into JSON structure
	uj, _ := json.Marshal(soundFileInfo)

	return httpresponse.CreateSuccessResponse(&cc.Context, http.StatusCreated, "Sound file upload success", "Sound file uploaded successfully", uj)
}

//func (sfc *SoundFileInfoController) TextToSpeech(c echo.Context) error {
//	// Grab user id for which file is being uploaded
//	id := c.Param("id")
//
//	userInfo, errUser := sfc.UserService.UserObject(id)
//	if errUser != nil {
//		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
//	}
//
//	// check if user is authorized to make this change
//	//if !userInfo.Role.IsAuthorized() {
//	//	return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UnauthorizedForAction)
//	//}
//
//	// Stub an user to be populated from the body
//	input := TTSRequest{}
//
//	if err := c.Bind(&input); err != nil {
//		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.SoundFileInputFormatIncorrect)
//	}
//
//	// validate input request body
//	if err := c.Validate(input); err != nil {
//		logrus.Infoln(err)
//		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
//	}
//
//	dest, duration,size ,err := sfc.TTSService.Convert(input.Text, input.Gender)
//	if err != nil {
//		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
//	}
//
//	// process sound file upload for user
//	soundFileInfo, err1 := sfc.processSoundFile(userInfo, dest, duration,size)
//	if err1 != nil {
//		cmlutils.DeleteFile(dest)
//		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, err1.Error())
//	}
//
//	// Marshal provided interface into JSON structure
//	uj, _ := json.Marshal(soundFileInfo)
//
//	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Sound file upload success", "Sound file uploaded successfully", uj)
//}

func (sfc *SoundFileInfoController) GetInfo(c echo.Context) error {
	// Grab user id for which file is being uploaded
	soundFileInfoId := c.Param("sid")

	soundFileInfo, er2 := sfc.SoundFileInfoService.SoundFileObjectWithId(soundFileInfoId)
	if er2 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.SoundFileDoesNotExist)
	}

	// Marshal provided interface into JSON structure
	sf, _ := json.Marshal(soundFileInfo)

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Name update", "Name updated", sf)
}

// this method is to update sound file name
// initially files name is set as name
// later it can be changed using this method
func (sfc *SoundFileInfoController) SoundFileNameUpdate(c echo.Context) error {

	// Grab user id for which file is being uploaded
	userID := c.Param("id")
	soundFileInfoId := c.Param("sid")

	// Stub an user to be populated from the body
	soundFileNameObject := NameUpdateObject{}

	if err := c.Bind(&soundFileNameObject); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.SoundFileNameUpdateFormatIncorrect)
	}

	// validate input request body
	if err := c.Validate(soundFileNameObject); err != nil {
		logrus.Infoln(err)
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	// validate if contact name is empty
	if soundFileNameObject.Name == "" {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.SoundFileUpdateNameMissing)
	}

	user, er1 := sfc.UserService.UserObject(userID)
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	soundFileInfo, er2 := sfc.SoundFileInfoService.SoundFileObjectWithId(soundFileInfoId)
	if er2 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.SoundFileDoesNotExist)
	}

	if user.ID.Hex() != soundFileInfo.UserID.Hex() {
		// given user is not authorized to make this call
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.SoundFileNoAccessToUpdate)
	}

	// update name for sound file
	soundFileInfo.Name = soundFileNameObject.Name

	// save update in db
	if err := sfc.SoundFileInfoService.SaveSoundFile(soundFileInfo); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	// Marshal provided interface into JSON structure
	sf, _ := json.Marshal(soundFileInfo)

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Name update", "Name updated", sf)
}

func (sfc *SoundFileInfoController) SoundFileDeleteMultiple(c echo.Context) error {
	cc := c.(model.AppContext)

	// Grab user id for which file is being uploaded
	//soundFileInfoId := c.Param("sid")
	soundFileNameObject := SoundFileIDs{}
	var count int
	if err := c.Bind(&soundFileNameObject); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	for _, id := range soundFileNameObject.ID {

		soundFileInfo, er2 := sfc.SoundFileInfoService.SoundFileObjectWithId(id)
		if er2 != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.SoundFileDoesNotExist)
		}

		bool, err := sfc.CampaignService.IsSoundFileAttached(cc.ID.Hex(), id)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
		}

		if bool {
			continue
			// contact list is attached to some campaign, you can not delete this contact list
			//return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.SoundFileAttachedWithCamapign)
		}

		if !sfc.soundFileExistsOnS3(*soundFileInfo) {
			// sound file doesn't exist on server
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.SoundFileDeleteFailForS3)
		}

		// delete file from s3
		sfc.deleteSoundFileFromS3(*soundFileInfo)

		if err := sfc.SoundFileInfoService.DeleteSoundFile(soundFileInfo); err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.SoundFileDeleteOperationFailed)
		}
		count++
		//	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Sound File Deleted successfully", "Sound File Deleted successfully")
	}
	submessage := fmt.Sprintf("%d Sound File Deleted successfully", count)
	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Sound File Deleted successfully", submessage)
}

// sound files can be deleted using this method
func (sfc *SoundFileInfoController) SoundFileDelete(c echo.Context) error {
	cc := c.(model.AppContext)

	// Grab user id for which file is being uploaded
	soundFileInfoId := c.Param("sid")

	soundFileInfo, er2 := sfc.SoundFileInfoService.SoundFileObjectWithId(soundFileInfoId)
	if er2 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.SoundFileDoesNotExist)
	}

	bool, err := sfc.CampaignService.IsSoundFileAttached(cc.ID.Hex(), soundFileInfoId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	if bool {
		// contact list is attached to some campaign, you can not delete this contact list
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.SoundFileAttachedWithCamapign)
	}

	if !sfc.soundFileExistsOnS3(*soundFileInfo) {
		// sound file doesn't exist on server
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.SoundFileDeleteFailForS3)
	}

	// delete file from s3
	sfc.deleteSoundFileFromS3(*soundFileInfo)

	if err := sfc.SoundFileInfoService.DeleteSoundFile(soundFileInfo); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.SoundFileDeleteOperationFailed)
	}

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Sound File Deleted successfully", "Sound File Deleted successfully")
}

// get list of all sound file
// this method also supports server side pagination
func (sfc *SoundFileInfoController) SoundFilesGetList(c echo.Context) error {
	cc := c.(model.AppContext)
	// Grab user id for which file is being uploaded
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")

	pageNumber, errPage := strconv.Atoi(page)
	limitNumber, errLimit := strconv.Atoi(limit)

	if errPage != nil || errLimit != nil {
		pageNumber = 1
		limitNumber = 10
	}

	if pageNumber < 1 {
		pageNumber = 1
	}

	if limitNumber < 1 {
		limitNumber = 10
	}

	soundFileList, count, err := sfc.SoundFileInfoService.SoundFileListWithUserId(cc.ID.Hex(), pageNumber, limitNumber)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	pageInfo := model.Page{}
	pageInfo.Page = pageNumber
	pageInfo.Limit = limitNumber
	pageInfo.TotalCount = count

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(soundFileList)
	pageData, _ := json.Marshal(pageInfo)

	return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusOK, "Sound files fetched successfully", "Sound files fetched successfully", data, pageData)
}

// get list of all default sound files
func (sfc *SoundFileInfoController) DefaultSoundFiles(c echo.Context) error {
	list, err := sfc.DefaultSoundFileInfoService.SoundFiles()
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(list)
	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Default Audios", "Default Audios", data)
}

// private method
// this method process the sound file present at the given destination
func (sfc *SoundFileInfoController) processSoundFile(filename string, userInfo *model.User, name string, originalFilePath string, duration int, size int, filetype string) (*model.SoundFileInfo, error) {

	soundFileInfo := sfc.CreateSoundFileInfoObject(userInfo, originalFilePath)
	soundFileInfo.Duration = duration
	soundFileInfo.FileSize = strconv.Itoa(size)
	soundFileInfo.FileType = filetype
	soundFileInfo.FileName = filename
	if len(name) > 0 {
		soundFileInfo.Name = name
	}

	error, wavFilePath, pcmaFilePath := sfc.SoundFileInfoService.PerformSoundFileConversions(originalFilePath)
	if error != nil {
		return nil, error
	}

	logrus.Infoln("local original file path : " + originalFilePath + " ̰")
	logrus.Infoln("local wav file path : " + wavFilePath + "\n")
	logrus.Infoln("local pcma file path : " + pcmaFilePath + "\n")

	// destination path for s3
	destinationPath := "/account/" + userInfo.ID.Hex() + "/soundfile/" + soundFileInfo.ID.Hex() + "/"

	error1, originalFileS3Path, wavFileS3Path, pcmaFileS3Path := sfc.uploadSoundFilesToS3(destinationPath, originalFilePath, wavFilePath, pcmaFilePath)
	if error1 != nil {
		// delete all temporary files
		//	sfc.deleteLocalSoundFiles(originalFilePath, wavFilePath, pcmaFilePath)
		return nil, error1
	}

	logrus.Infoln("original file path s3 : " + originalFileS3Path + "\n")
	logrus.Infoln("wav file path s3 : " + wavFileS3Path + "\n")
	logrus.Infoln("pcma file path s3 : " + pcmaFileS3Path + "\n")

	soundFileInfo.OriginalFilePath = originalFileS3Path
	soundFileInfo.WavFilePath = wavFileS3Path
	soundFileInfo.PCMAFilePath = pcmaFileS3Path

	// delete all temporary files
	sfc.deleteLocalSoundFiles(originalFilePath, wavFilePath, pcmaFilePath)

	if err := sfc.SoundFileInfoService.SaveSoundFile(soundFileInfo); err != nil {
		// delete uploaded files on s3 as well
		sfc.S3Service.DeleteS3Object(cmlutils.S3BucketName(), originalFileS3Path)
		sfc.S3Service.DeleteS3Object(cmlutils.S3BucketName(), wavFileS3Path)
		sfc.S3Service.DeleteS3Object(cmlutils.S3BucketName(), pcmaFileS3Path)
		return nil, err
	}

	return soundFileInfo, nil
}

// private method
// this method saves given to a destination path
// it returns destination file path in case of success
// in case of failure it returns error
func (sfc *SoundFileInfoController) saveMultipartFile(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", errors.New(cmlmessages.SoundFileReadError)
	}
	defer src.Close()

	fileName, err := formatFileName(file.Filename)
	if err != nil {
		return "", err
	}

	// Destination
	destinationPath := cmlconstants.TempDestinationPath + fileName

	// prepare the dst
	os.MkdirAll(cmlconstants.TempDestinationPath, os.ModePerm)

	dst, err := os.Create(destinationPath)
	if err != nil {
		return "", errors.New(cmlmessages.SoundFileWriteError)
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return "", errors.New(cmlmessages.SoundFileWriteError)
	}

	return destinationPath, nil
}

func formatFileName(src string) (string, error) {
	basePath := filepath.Base(src)
	ext := filepath.Ext(src)
	fileName := strings.TrimSuffix(basePath, ext)

	fn, err := cmlutils.RemoveNonAlphaNumeric(fileName)
	if err != nil {
		return "", errors.New("File name is invalid")
	}

	fileName = fn + ext
	return fileName, nil
}

// private method
// it deletes all given files from local server
func (sfc *SoundFileInfoController) deleteLocalSoundFiles(originalFile string,
	wavFilePath string,
	pcmaFilePath string) {

	cmlutils.DeleteFile(wavFilePath)
	cmlutils.DeleteFile(pcmaFilePath)
	cmlutils.DeleteFile(originalFile)
}

func (sfc *SoundFileInfoController) uploadSoundFilesToS3(destinationPath string,
	originalFilePath string,
	wavFilePath string,
	pcmaFilePath string) (error, string, string, string) {

	// upload original file to s3
	originalFileS3Path, err := sfc.S3Service.Upload(destinationPath, originalFilePath)
	if err != nil {
		return errors.New("File Upload Error"), "", "", ""
	}

	// upload wav file to s3
	wavFileS3Path, err := sfc.S3Service.Upload(destinationPath, wavFilePath)
	if err != nil {
		// delete already uploaded files
		sfc.S3Service.DeleteS3Object(cmlutils.S3BucketName(), originalFileS3Path)
		return errors.New("File Upload Error"), "", "", ""
	}

	// upload pcma file to s3
	pcmaFileS3Path, err := sfc.S3Service.Upload(destinationPath, pcmaFilePath)
	if err != nil {
		// delete already uploaded files
		sfc.S3Service.DeleteS3Object(cmlutils.S3BucketName(), originalFileS3Path)
		sfc.S3Service.DeleteS3Object(cmlutils.S3BucketName(), wavFileS3Path)
		return errors.New("File Upload Error"), "", "", ""
	}

	return nil, originalFileS3Path, wavFileS3Path, pcmaFileS3Path
}

func (sfc *SoundFileInfoController) CreateSoundFileInfoObject(user *model.User,
	originalFilePath string) *model.SoundFileInfo {

	fileBaseName := filepath.Base(originalFilePath)
	fileName := strings.TrimSuffix(fileBaseName, filepath.Ext(fileBaseName))

	soundFileInfo := model.SoundFileInfo{}
	soundFileInfo.ID = bson.NewObjectId()
	soundFileInfo.Name = fileName
	soundFileInfo.FileName = fileBaseName
	soundFileInfo.UserID = user.ID
	soundFileInfo.SoundFileCode = cmlutils.EpochMilli()

	return &soundFileInfo
}

// private method
// SoundFileExistsOnS3 checks if the given sound file object
// exists in s3 or not
// returns boolean value
func (sfc *SoundFileInfoController) soundFileExistsOnS3(soundFileObject model.SoundFileInfo) bool {
	// check if original file exists
	if err := sfc.S3Service.CheckIfObjectExists(cmlutils.S3BucketName(), soundFileObject.OriginalFilePath); err != nil {
		return false
	}

	// check if .wav file exists
	if err := sfc.S3Service.CheckIfObjectExists(cmlutils.S3BucketName(), soundFileObject.WavFilePath); err != nil {
		return false
	}

	// check if .pcma file exists
	if err := sfc.S3Service.CheckIfObjectExists(cmlutils.S3BucketName(), soundFileObject.PCMAFilePath); err != nil {
		return false
	}

	return true
}

// private method
// this method deleted all resouce files for a particular sound file on s3
func (sfc *SoundFileInfoController) deleteSoundFileFromS3(soundFileInfo model.SoundFileInfo) {

	// delete original file
	if err := sfc.S3Service.DeleteS3Object(cmlutils.S3BucketName(), soundFileInfo.OriginalFilePath); err != nil {
		//TODO: log this delete fail in a seperate table for .wav
		logrus.Infoln("Delete failed for original file")
	}

	// delete .wav file
	if err := sfc.S3Service.DeleteS3Object(cmlutils.S3BucketName(), soundFileInfo.WavFilePath); err != nil {
		//TODO: log this delete fail in a seperate table for .wav
		logrus.Infoln("Delete failed for .wav file")
	}

	// delete .pcma file
	if err := sfc.S3Service.DeleteS3Object(cmlutils.S3BucketName(), soundFileInfo.PCMAFilePath); err != nil {
		//TODO: log this delete fail in a seperate table for .pcma
		logrus.Infoln("Delete failed for .PCMA file")
	}
}
