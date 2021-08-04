package service

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
)

type S3Service struct {
}

func NewS3Service() *S3Service {
	return &S3Service{}
}

func (s3s *S3Service) Upload(destinationPath string, sourcePath string) (string, error) {

	// The session the S3 Uploader will use
	region := os.Getenv("AWS_REGION")
	session := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(session)

	f, err := os.Open(sourcePath)
	if err != nil {
		fmt.Errorf("failed to open file %q, %v", sourcePath, err)
		return "", err
	}

	defer f.Close()

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(cmlutils.S3BucketName() + destinationPath),
		Key:    aws.String(filepath.Base(sourcePath)),
		Body:   f,
	})

	if err != nil {
		logrus.Errorln(err.Error())
		fmt.Errorf("failed to upload file %q, %v", filepath.Base(sourcePath), err)
		return "", err
	}

	logrus.Infoln("upload location ", result.Location)
	return destinationPath + filepath.Base(sourcePath), err
}

func (s3s *S3Service) DownloadFile(filepath string, url string) error {

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

func (s3s *S3Service) DeleteS3Object(bucket string, key string) error {
	// Initialize a session that the SDK uses to load configuration,
	// credentials, and region from the shared config file. (~/.aws/config).
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create S3 service client
	svc := s3.New(session)

	// Delete the item
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})

	if err != nil {
		logrus.Errorln("s3 object delete error : " + err.Error())
		return err
	}

	return err
}

func (s3s *S3Service) CheckIfObjectExists(bucket string, key string) error {
	// Initialize a session that the SDK uses to load configuration,
	// credentials, and region from the shared config file. (~/.aws/config).
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create S3 service client
	svc := s3.New(session)

	headInput := s3.HeadObjectInput{Bucket: &bucket, Key: &key}

	output, err := svc.HeadObject(&headInput)
	logrus.Infoln(output)
	if err != nil {
		return err
	}
	return nil
}
