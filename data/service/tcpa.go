package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type TCPAService struct {
}

func NewTCPAService() *TCPAService {
	return &TCPAService{}
}

type TCPARequest struct {
	Type        string `json:"type"`
	PhoneNumber string `json:"phones"`
}

type TCPAStatusRequest struct {
	Key string `json:"key"`
}

type TCPANumberRequest struct {
	Phones string   `json:"phones"`
	Type   []string `json:"type"`
}

type TCPAJobStatusResponse struct {
}

func (tcpas *TCPAService) IsNumberInTCPA(numberlist []string, meta map[string]interface{}) (string, error) {

	apiurl := "https://api.tcpalitigatorlist.com/scrub/phones"

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	result1 := strings.Join(numberlist, ",")
	numerbs := "[" + result1 + "]"

	body := TCPANumberRequest{
		Phones: numerbs,
		Type:   []string{"tcpa", "dnc", "dnc_fed"},
	}

	byteBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	username := os.Getenv("TCPA_USERNAME")
	passwordorg := "1qm2 gVv4 uUgL 6TFz 4m33 Qalv"

	req, err := http.NewRequest("POST", apiurl, bytes.NewBuffer(byteBody))
	if err != nil {
		logrus.Errorln(fmt.Sprintf("IsNumberInTCPA  =>  MetaData : %s  Request Body : %s  :: Error :%s ::", meta, string(byteBody), err.Error()))
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(username, passwordorg)

	resp, err2 := client.Do(req)
	if err2 != nil {
		logrus.Errorln(fmt.Sprintf("IsNumberInTCPA => MetaData : %s Request Header : %s ::Request Url  :%s   :: Error :%s ::", meta, req.Header, req.URL, err2.Error()))
		return "", err2
	}
	if resp.StatusCode != http.StatusOK {
		logrus.Errorln(fmt.Sprintf("IsNumberInTCPA=>  MetaData : %s  Request Header : %s :: Request Response :%s :: Request Url :%s :: Request Status Code :%d :: Error :%s ::", meta, req.Header, req.Response.Body, req.URL, resp.StatusCode, err2.Error()))
		return "", errors.New(resp.Status)
	}

	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if _, ok := result["job_key"]; ok {
		return result["job_key"].(string), nil
	}
	return "", nil
}

func (tcpas *TCPAService) CheckTCPAJobStatus(key string, m map[string]interface{}) (map[string]interface{}, error) {
	client := &http.Client{}
	var jobkey TCPAStatusRequest
	jobkey.Key = key
	b, err := json.Marshal(jobkey)
	reqbody := []byte(b)
	req, err := http.NewRequest("POST", "https://api.tcpalitigatorlist.com/scrub/phones/get", bytes.NewBuffer(reqbody))
	username := os.Getenv("TCPA_USERNAME")
	//password := os.Getenv("TCPA_PASSWORD")
	passwordorg := "1qm2 gVv4 uUgL 6TFz 4m33 Qalv"

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(username, passwordorg)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {

		logrus.Errorln(fmt.Sprintf("CheckTCPAJobStatus  => MetaData : %s :: Request Body : %s :: Error :%s ::", m, req.Body, err.Error()))
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		logrus.Errorln(fmt.Sprintf("CheckTCPAJobStatus  =>  MetaData : %s :: Request Body : %s :: Error :%s ::", m, req.Body, err.Error()))
		return nil, err
	}

	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

var dnclistt []string

func parseMap(aMap map[string]interface{}) (dnclist []string) {
	for key, val := range aMap {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			if key == "match" {
				parseMap(val.(map[string]interface{}))
			}
			dnclistt = append(dnclistt, key)
		case []interface{}:
			parseArray(val.([]interface{}))
		default:
			logrus.Errorln(key, ":", concreteVal)
		}
	}

	return dnclistt
}

func parseArray(anArray []interface{}) {
	for i, val := range anArray {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			logrus.Errorln("Index:", i)
			parseMap(val.(map[string]interface{}))
		case []interface{}:
			logrus.Errorln("Index:", i)
			parseArray(val.([]interface{}))
		default:
			logrus.Errorln("Index", i, ":", concreteVal)
		}
	}
}
