package service

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

const TempDestinationPath = "./Temp/"

func NewFileService() *FileService {
	return &FileService{}
}

type FileService struct {
}

func (f *FileService) IsValidImageFileName(name string) bool {
	isValidFileFormat := false

	if strings.HasSuffix(name, ".png") {
		isValidFileFormat = true
	}

	if strings.HasSuffix(name, ".jpg") {
		isValidFileFormat = true
	}

	if strings.HasSuffix(name, ".jpeg") {
		isValidFileFormat = true
	}

	return isValidFileFormat
}

func (f *FileService) SaveFile(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Destination
	destinationPath := TempDestinationPath + file.Filename
	destinationPath = strings.Replace(destinationPath, " ", "_", -1)

	// prepare the dst
	os.MkdirAll(TempDestinationPath, os.ModePerm)

	logrus.Infoln(destinationPath)
	dst, err := os.Create(destinationPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	return destinationPath, nil
}

func (f *FileService) DeleteFile(path string) error {
	// delete file
	return os.Remove(path)
}

func (f *FileService) DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func (f *FileService) MerageFile(name string, LocalFile string, S3File string) (path string, err error) {
	out, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logrus.Errorln("failed to open outpout file:", err)
		return "", err
	}
	defer out.Close()
	zipIn, err := os.Open(LocalFile)
	if err != nil {
		logrus.Errorln("failed to open outpout file:", err)
		return "", err
	}
	defer zipIn.Close()

	signedIn, err := os.Open(S3File)
	if err != nil {
		logrus.Errorln("failed to open outpout file:", err)
		return "", err
	}
	defer signedIn.Close()
	_, err = io.Copy(out, zipIn)
	if err != nil {
		logrus.Errorln("failed to append zip file to output:", err)
		return "", err
	}

	_, err = io.Copy(out, signedIn)
	if err != nil {
		logrus.Errorln("failed to append signed file to output:", err)
		return "", err
	}
	return out.Name(), nil
}
