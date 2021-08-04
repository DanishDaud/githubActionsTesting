package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/jasonlvhit/gocron"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type NumberService struct {
	Service
	UserService
}

func (telnyx *NumberService) SearchBulkVSNumbers(op model.BulkVsSearchRequest) (*model.BulkVSSearchResponse, error) {

	client := &http.Client{}
	npa := op.Npa

	payload := url.Values{}

	payload.Add("Npa", npa)
	req, err := http.NewRequest("GET", "https://portal.bulkvs.com/api/v1.0/orderTn?"+payload.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	username := os.Getenv("BULK_VS_USERNAME")
	password := os.Getenv("BULK_VS_PASSWORD")
	req.SetBasicAuth(username, password)
	var result model.BulkVSSearchResponse
	resp, err := client.Do(req)
	if err != nil {
		return &result, err
	}
	if resp.StatusCode != http.StatusOK {
		return &result, errors.New("Please Try Later")
	}
	body, _ := ioutil.ReadAll(resp.Body)

	json.Unmarshal(body, &result)
	if result == nil {
		bodyString := string(body)
		return nil, errors.New(bodyString)
	}
	return &result, nil

}

func (telnyx *NumberService) PurchaseBulkVSNumbers(op model.BulkvsPurchaseNumbersRequest) ([]*model.TelnyxNumber, float32, error) {
	var objects []*model.TelnyxNumber
	client := &http.Client{}
	var numberobject model.BulkVsApiPurchaseRequest
	for i, _ := range op.Numbers {
		numberobject.TN = op.Numbers[i]
		numberobject.TrunkGroup = "fsinbound"
		numberobject.Lidb = ""
		numberobject.PortoutPin = "20978432"
		numberobject.Sms = true
		numberobject.Mms = false
		b, err := json.Marshal(numberobject)
		if err != nil {
			return nil, 0, err
		}
		reqbody := []byte(b)

		req, err := http.NewRequest("POST", "https://portal.bulkvs.com/api/v1.0/orderTn", bytes.NewBuffer(reqbody))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")
		username := os.Getenv("BULK_VS_USERNAME")
		password := os.Getenv("BULK_VS_PASSWORD")
		req.SetBasicAuth(username, password)
		if err != nil {
			return nil, 0, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, 0, err
		}
		var resultt model.BulkVsApiPurchaseResponse
		if resp.StatusCode != http.StatusOK {
			return nil, 0, errors.New("Please Try Later")
		}
		body, _ := ioutil.ReadAll(resp.Body)

		json.Unmarshal(body, &resultt)
		var number model.TelnyxNumber
		number.UserID = op.UserID
		number.PhoneNumber = numberobject.TN
		number.ConnectionID = numberobject.TrunkGroup
		number.PhoneStatus = resultt.Status

		result, err1 := telnyx.StoreNumber(&number)
		if err1 != nil {
			return nil, 0, err
		}
		objects = append(objects, result)

	}

	initial_rate, monthly_rate, err := telnyx.getBillingRate(op.UserID.Hex())
	if err != nil {
		return nil, 0, err
	}
	count := len(op.Numbers)
	var countt float32 = float32(count)

	amount := countt * (initial_rate + monthly_rate)

	return objects, amount, nil
}

func (telnyx *NumberService) getBillingRate(userId string) (float32, float32, error) {

	userDataSource := telnyx.userDataSource()
	defer userDataSource.Session.Close()
	id := bson.ObjectIdHex(userId)
	user, err := userDataSource.GetUserById(id)
	if err != nil {
		return -1, 0.0, err
	}
	parentUser, err := userDataSource.GetUserById(user.ParentID)
	initialrate := parentUser.Defaults.InitialPurchaseRate
	monthlyrate := parentUser.Defaults.MonthlyPurchaseRate

	if initialrate == 0.0 {
		return -1, 0.0, errors.New("billing rate can not be null")
	}

	return initialrate, monthlyrate, err
}

func (telnyx *NumberService) DeleteBulkVSSingleNumber(id string) error {

	if err := !bson.IsObjectIdHex(id); err != false {
		return errors.New("invalid Object ID")
	}
	object, err := telnyx.GetNumberByID(bson.ObjectIdHex((id)))
	if err != nil {
		return err
		// TODO : REMOVE THIS and log it
	}
	client := &http.Client{}
	payload := url.Values{}
	payload.Add("Number", object.PhoneNumber)
	req, err := http.NewRequest("DELETE", "https://portal.bulkvs.com/api/v1.0/tnRecord?"+payload.Encode(), nil)

	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	username := os.Getenv("BULK_VS_USERNAME")
	password := os.Getenv("BULK_VS_PASSWORD")
	req.SetBasicAuth(username, password)
	req.SetBasicAuth("smartautomation", "b8719066674f74b9c52f0ed7f944e7be")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("Error Deleting Number")
	}

	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)
	//_ , _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("response Body:", string(body))
	//myText := strings.Replace(id, "-", "", 1)
	//myText = strings.TrimSpace(id)
	//fmt.Println(myText)
	obj, err := telnyx.GetNumberCarrierGroupByNumberID(object.ID)
	if obj == nil {
		telnyx.DeleteNumberByID(object.ID)
		return err
	}
	for i, _ := range obj {
		err2 := telnyx.DeleteFromContactGroup(obj[i].ContactGroupID, object.PhoneNumber)
		if err2 != nil {
			return err2
		}
		err3 := telnyx.DeleteFromNumberCarrierGroup(obj[i].ID)
		if err != nil {
			return err3
		}
		err1 := telnyx.DeleteNumberByID(object.ID)
		if err1 != nil {
			return err1
		}
	}

	return nil
}

func (telnyx *NumberService) SearchTelnyxNumbers(op model.TelnyxSearchNumbersRequest) (*model.TelnyxSearchResponse, error) {

	client := &http.Client{}
	countrycode := "us"
	locality := op.Locality
	number_type := op.NumberType
	administrivearea := op.AdministriveArea
	limit := op.Limit
	npa := op.Npa

	payload := url.Values{}
	if countrycode != "" {
		payload.Add("filter[country_code]", countrycode)
	}
	if administrivearea != "" {
		payload.Add("filter[administrative_area]", administrivearea)
	}
	if limit != "" {
		payload.Add("filter[limit]", limit)
	}
	if number_type != "" {
		payload.Add("filter[number_type]", number_type)
	}
	if locality != "" {
		payload.Add("filter[locality]", locality)
	}
	payload.Add("filter[national_destination_code]", npa)
	req, err := http.NewRequest("GET", "https://api.telnyx.com/v2/available_phone_numbers?"+payload.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	telnyx_key := "Bearer " + os.Getenv("TELNYX_API_KEY")
	req.Header.Add("Authorization", telnyx_key)

	var result model.TelnyxSearchResponse
	resp, err := client.Do(req)
	if err != nil {
		return &result, err
	}
	if resp.StatusCode != http.StatusOK {
		return &result, errors.New("Please Try Later")
	}
	body, _ := ioutil.ReadAll(resp.Body)

	json.Unmarshal(body, &result)
	return &result, nil

}

func (telnyx *NumberService) DeleteSingleNumber(id string) error {

	if err := !bson.IsObjectIdHex(id); err != false {
		return errors.New("invalid Object ID")
	}
	object, err := telnyx.GetNumberByID(bson.ObjectIdHex((id)))
	if err != nil {
		return err
		// TODO : REMOVE THIS and log it
	}
	client := &http.Client{}
	url := "https://api.telnyx.com/v2/phone_numbers/" + ("+1" + object.PhoneNumber)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	telnyx_key := "Bearer " + os.Getenv("TELNYX_API_KEY")
	req.Header.Add("Authorization", telnyx_key)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("Error Deleting Number")
	}
	logrus.Infoln("response Status:", resp.Status)
    logrus.Infoln("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Infoln("response Body:", string(body))
	//myText := strings.Replace(id, "-", "", 1)
	//myText = strings.TrimSpace(id)
	//fmt.Println(myText)
	obj, err := telnyx.GetNumberCarrierGroupByNumberID(object.ID)
	if err != nil {
		return err
	}
	for i, _ := range obj {
		err2 := telnyx.DeleteFromContactGroup(obj[i].ContactGroupID, object.PhoneNumber)
		if err2 != nil {
			return err2
		}
		err3 := telnyx.DeleteFromNumberCarrierGroup(obj[i].ID)
		if err != nil {
			return err3
		}
		err1 := telnyx.DeleteNumberByID(object.ID)
		if err1 != nil {
			return err1
		}
	}

	return nil
}

func (telnyx *NumberService) DeleteMultipleBulkVSNumbers(op model.DeleteNumbersRequest) error {
	for i, _ := range op.NumberID {
		if err := !bson.IsObjectIdHex(op.NumberID[i]); err != false {
			return errors.New("invalid Object ID")
		}
		err := telnyx.DeleteBulkVSSingleNumber(op.NumberID[i])
		if err != nil {
			return err
		}

	}
	return nil
}

func (telnyx *NumberService) PurchaseTelnyxNumbers(op model.PurchaseNumbersRequest) ([]*model.TelnyxNumber, error) {
	var objects []*model.TelnyxNumber
	client := &http.Client{}
	box := []model.BaseTelnyxPurchaseRequest{}
	for _, v := range op.Numbers {
		box = append(box, model.BaseTelnyxPurchaseRequest{v})
	}

	boxx := model.FinalTelnyxPurchaseRequest{}
	boxx.Numbers = box

	b, err := json.Marshal(boxx)
	if err != nil {
		return nil, err
	}

	reqbody := []byte(b)

	req, err := http.NewRequest("POST", "https://api.telnyx.com/v2/number_orders", bytes.NewBuffer(reqbody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	telnyx_key := "Bearer " + os.Getenv("TELNYX_API_KEY")
	req.Header.Add("Authorization", telnyx_key)

	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Please Try Later")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Infoln("response Body:", string(body))
	var result model.TelnyxPurchaseResponse
	json.Unmarshal(body, &result)

	for i, _ := range op.Numbers {
		var number model.TelnyxNumber
		number.PhoneID = string(result.Data.PhoneNumbers[i].ID)
		number.PhoneNumber = (strings.Replace(result.Data.PhoneNumbers[i].PhoneNumber, "+1", "", 1))
		number.PhoneStatus = result.Data.PhoneNumbers[i].Status
		number.TotalInbound = 0
		number.TotalOutbound = 0
		number.InboundDuration = 0

		connection_id := os.Getenv("TELNYX_CONNECTION_ID")
		message := `{"connection_id": "` + connection_id + `"}`
		reqbody := []byte(message)
		url := "https://api.telnyx.com/v2/phone_numbers/" + result.Data.PhoneNumbers[i].PhoneNumber
		req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(reqbody))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")
		telnyx_key := "Bearer " + os.Getenv("TELNYX_API_KEY")
		req.Header.Add("Authorization", telnyx_key)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("Error Attaching Connection to Number")
		}

		defer resp.Body.Close()
		logrus.Infoln("response Status:", resp.Status)
		logrus.Infoln("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		logrus.Infoln("response Body:", string(body))
		number.ConnectionID = os.Getenv("TELNYX_CONNECTION_ID")
		number.UserID = op.UserID
		result, err1 := telnyx.StoreNumber(&number)
		if err1 != nil {
			return nil, err1
		}
		objects = append(objects, result)

	}
	gocron.Every(15).Second().Do(telnyx.CheckStatusTelnyxNumbers, op)

	return objects, err
}

func (telnyx *NumberService) CheckStatusTelnyxNumbers(op model.PurchaseNumbersRequest) error {
	client := &http.Client{}
	for i, _ := range op.Numbers {
		var number *model.TelnyxNumber
		//number.OrderID = string(result.Data.ID)
		number, err := telnyx.GetTelnyxNumber((strings.Replace(op.Numbers[i], "+1", "", 1)))
		if err != nil {
			return err
		}
		modifiednumber := "+1" + number.PhoneNumber
		url := "https://api.telnyx.com/v2/phone_numbers/" + modifiednumber
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")
		telnyx_key := "Bearer " + os.Getenv("TELNYX_API_KEY")
		req.Header.Add("Authorization", telnyx_key)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New("Carrier Status Error")
		}
		defer resp.Body.Close()
		logrus.Infoln("response Status:", resp.Status)
		logrus.Infoln("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		logrus.Infoln("response Body:", string(body))

		var result model.TelnyxStatusResponse
		json.Unmarshal(body, &result)

		if result.Data.Status == "active" {
			gocron.Remove(telnyx.CheckStatusTelnyxNumbers)
			gocron.Clear()
			//number.PhoneNumber= result.Data.PhoneNumber
			number.PhoneStatus = result.Data.Status
			//fmt.Println(result.Data.Status)
			//fmt.Println(result.Data.PhoneNumber)
			_, err1 := telnyx.UpdateByStatus(number)
			if err1 != nil {
				return err1
			}
		}

	}
	return nil
}

func (uc *NumberService) GetForwardingNumber(number string) (string, error) {
	if len(number) > 10 {
		s := len(number) - 10
		number = number[s:]
	}
	telnyxDataSource := uc.telnyxDataSource()
	defer telnyxDataSource.Session.Close()
	forwardnumber, err := telnyxDataSource.GetForwardingNumberByNumber(number)
	if err != nil {
		return "", err
	}

	return forwardnumber, nil
}

func (telnyx *NumberService) GetAll(op model.GetAllRequest) (*model.TelnyxList, int, error) {

	telnyxDataSource := telnyx.telnyxDataSource()
	defer telnyxDataSource.Session.Close()
	list, total, err := telnyxDataSource.GetAll(op)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (telnyx *NumberService) GetTelnyxNumber(obj string) (*model.TelnyxNumber, error) {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxDataSource()
	defer tnd.Session.Close()
	return tnd.GetByNumber(obj)
}

func (telnyx *NumberService) GetNumberByID(id bson.ObjectId) (*model.TelnyxNumber, error) {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxDataSource()
	defer tnd.Session.Close()
	return tnd.GetByID(id)
}

func (telnyx *NumberService) StoreNumber(obj *model.TelnyxNumber) (*model.TelnyxNumber, error) {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxDataSource()
	defer tnd.Session.Close()
	return tnd.Store(obj)
}

func (telnyx *NumberService) DeleteTelnyxNumber(num string) error {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxDataSource()
	defer tnd.Session.Close()
	id, err := tnd.GetByNumber(num)
	if err != nil {
		return err
	}
	return tnd.DeleteById(id.ID)
}

func (telnyx *NumberService) DeleteNumberByID(id bson.ObjectId) error {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxDataSource()
	defer tnd.Session.Close()
	return tnd.DeleteById(id)
}

func (telnyx *NumberService) DeleteFromContactGroup(id bson.ObjectId, num string) error {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxDataSource()
	defer tnd.Session.Close()
	return tnd.DeleteTelnyxNumberFromContactGroup(id, num)
}

func (telnyx *NumberService) DeleteFromNumberCarrierGroup(id bson.ObjectId) error {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxDataSource()
	defer tnd.Session.Close()
	return tnd.DeleteTelnyxGroupByID(id)
}

func (telnyx *NumberService) GetNumberCarrierGroupByNumberID(id bson.ObjectId) ([]model.TelnyxNumbersGroup, error) {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxDataSource()
	defer tnd.Session.Close()
	return tnd.GetNumberCarrierGroupByNumberID(id)
}

func (telnyx *NumberService) UpdateByStatus(obj *model.TelnyxNumber) (*model.TelnyxNumber, error) {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxDataSource()
	defer tnd.Session.Close()
	return tnd.UpdateByStatus(obj)
}

func (telnyx *NumberService) UpdateSingle(object model.TelnyxNumber) (*model.TelnyxNumber, error) {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxDataSource()
	defer tnd.Session.Close()
	return tnd.UpdateByForwardingNumber(&object)
}

func (telnyx *NumberService) UpdateSingleNumber(op model.UpdateSingleNumberRequest) (*model.TelnyxNumber, error) {
	var number model.TelnyxNumber
	//match, _ := regexp.MatchString(`^\\d{1,10}$`, op.ForwardingNumber)
	if err := !bson.IsObjectIdHex(op.NumberID); err != false {
		return nil, errors.New("invalid Object ID")
	}
	number.ID = bson.ObjectIdHex(op.NumberID)
	number.ForwardingNumber = op.ForwardingNumber
	result, err1 := telnyx.UpdateSingle(number)
	if err1 != nil {
		return nil, err1
	}
	return result, nil
}

func (telnyx *NumberService) UpdateMultipleNumbers(op model.UpdateMultipleNumberRequest) ([]*model.TelnyxNumber, error) {
	var objects []*model.TelnyxNumber
	for i, _ := range op.NumberID {
		var number model.TelnyxNumber
		if err := !bson.IsObjectIdHex(op.NumberID[i]); err != false {
			return nil, errors.New("invalid Object ID")
		}
		number.ID = bson.ObjectIdHex(op.NumberID[i])
		number.ForwardingNumber = op.ForwardingNumber
		result, err1 := telnyx.UpdateSingle(number)
		if err1 != nil {
			return nil, err1
		}
		objects = append(objects, result)

	}
	return objects, nil
}

func (telnyx *NumberService) DeleteMultipleNumbers(op model.DeleteNumbersRequest) error {
	for i, _ := range op.NumberID {

		if err := !bson.IsObjectIdHex(op.NumberID[i]); err != false {
			return errors.New("invalid Object ID")
		}
		obj, err := telnyx.GetNumberCarrierGroupByNumberID(bson.ObjectIdHex(op.NumberID[i]))
		if err != nil {
			return err
		}
		number, err1 := telnyx.GetNumberByID(bson.ObjectIdHex(op.NumberID[i]))
		if err1 != nil {
			return err1
		}
		err2 := telnyx.DeleteFromContactGroup(obj[i].ContactGroupID, number.PhoneNumber)
		if err2 != nil {
			return err2
		}
		err3 := telnyx.DeleteFromNumberCarrierGroup(obj[i].ID)
		if err3 != nil {
			return err3
		}
		err4 := telnyx.DeleteNumberByID(bson.ObjectIdHex(op.NumberID[i]))
		if err4 != nil {
			return err4
		}
	}
	return nil
}

func (telnyx *NumberService) telnyxDataSource() *datasource.NumberDataSource {
	return &datasource.NumberDataSource{DataSource: datasource.DataSource{Session: telnyx.Session.Copy()}}
}
