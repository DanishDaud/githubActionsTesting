package service

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlconstants"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
)

type TTSService struct {
}

func NewTTSService() *TTSService {
	return &TTSService{}
}

// returns destination path, file length, and error if any
func (tts *TTSService) Convert(text string, gender int) (string, int, int, error) {
	// aws session
	region := os.Getenv("AWS_REGION")
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))

	// Create Polly client
	svc := polly.New(sess)

	voiceId := "Brian"
	if gender == 2 {
		voiceId = "Joanna"
	}
	// Output to MP3 using voice Joanna
	input := &polly.SynthesizeSpeechInput{OutputFormat: aws.String("mp3"), Text: aws.String(text), VoiceId: aws.String(voiceId)}

	output, err := svc.SynthesizeSpeech(input)
	if err != nil {
		logrus.Errorln("Got error calling SynthesizeSpeech:")
		logrus.Errorln(err.Error())
		return "", 0, 0, err
	}

	// Save as MP3
	name := "tts_" + strconv.Itoa(int(cmlutils.RandInt(10000, 99999)))
	mp3File := cmlconstants.TempDestinationPath + name + ".mp3"

	outFile, err := os.Create(mp3File)
	if err != nil {
		logrus.Errorln("Got error creating " + mp3File + ":")
		logrus.Errorln(err.Error())
		return "", 0, 0, err
	}

	defer outFile.Close()
	_, err = io.Copy(outFile, output.AudioStream)
	if err != nil {
		logrus.Errorln("Got error saving MP3:")
		logrus.Errorln(err.Error())
		return "", 0, 0, err
	}

	duration, err := GetFileDuration(mp3File)
	if err != nil {
		return "", 0, 0, err
	}
	filesize, err := GetFileSize(mp3File)
	if err != nil {
		return "", 0, 0, err
	}
	return mp3File, duration, filesize, nil
}

func GetFileDuration(src string) (int, error) {
	// generate .wav file
	command := fmt.Sprintf("/usr/bin/ffprobe -i %s -show_entries format=duration -v quiet -of csv=\"p=0\"", src)

	logrus.Infoln(".wav conversion command : " + command)

	out, error1 := exec.Command("sh", "-c", command).Output()
	if error1 != nil {
		logrus.Errorln("Error finding file duration : ")
		logrus.Errorln(error1.Error())
		logrus.Infoln("\n")
		return 0, errors.New("Error converting file to .wav")
	}

	outString := string(out)
	outString = strings.TrimSuffix(outString, "\n")
	logrus.Infoln("Duration find output : ", outString)
	duration, err := strconv.ParseFloat(outString, 64)
	if err != nil {
		return 0, err
	}

	return int(duration), err
}
func GetFileSize(src string) (int, error) {
	file, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}

	var bytes int64
	bytes = stat.Size()

	var kilobytes int64
	kilobytes = (bytes / 1024)
	var kb int
	kb = int(kilobytes)
	logrus.Infoln("File size in kilobytes ", kilobytes)

	return kb, nil
}
