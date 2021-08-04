package service

import (
	"errors"
	"github.com/gomarkho/sas-rvm-provapi/data/datasource"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"gopkg.in/mgo.v2/bson"
)

type NumbersCarrierGroupService struct {
	Service
	NumberService
	ContactListService
	ContactGroupService
}

//
//
//type StatusResponseTelnyx struct {
//	Data struct {
//		ID                    string        `json:"id"`
//		RecordType            string        `json:"record_type"`
//		PhoneNumber           string        `json:"phone_number"`
//		Status                string        `json:"status"`
//		Tags                  []interface{} `json:"tags"`
//		ConnectionID          string        `json:"connection_id"`
//		ExternalPin           interface{}   `json:"external_pin"`
//		T38FaxGatewayEnabled  bool          `json:"t38_fax_gateway_enabled"`
//		PurchasedAt           time.Time     `json:"purchased_at"`
//		BillingGroupID        interface{}   `json:"billing_group_id"`
//		EmergencyEnabled      bool          `json:"emergency_enabled"`
//		CallForwardingEnabled bool          `json:"call_forwarding_enabled"`
//		CnamListingEnabled    bool          `json:"cnam_listing_enabled"`
//		CallRecordingEnabled  bool          `json:"call_recording_enabled"`
//		MessagingProfileID    string        `json:"messaging_profile_id"`
//		MessagingProfileName  string        `json:"messaging_profile_name"`
//		CreatedAt             time.Time     `json:"created_at"`
//		UpdatedAt             time.Time     `json:"updated_at"`
//	} `json:"data"`
//}
//
//
//
//
type TelnyxNumbersGroupCreateMultiple struct {
	NumberID      []string      `json:"numberID"`
	UserID        bson.ObjectId `bson:"userid"`
	ContactListID bson.ObjectId `json:"contactListID" bson:"contactListID"`
}

type TelnyxNumbersGroupCreateSingle struct {
	NumberID      string        `json:"numberID"`
	UserID        bson.ObjectId `bson:"userid"`
	ContactListID bson.ObjectId `json:"contactListID" bson:"contactListID"`
}

type TelnyxNumbersGroupDeleteMultipleRequest struct {
	NumberID      string `json:"numberID"`
	UserID        bson.ObjectId
	ContactListID []string `json:"contactListID" `
}

type TelnyxNumbersGroupDeleteSingleRequest struct {
	NumberID      string `json:"numberID"`
	UserID        bson.ObjectId
	ContactListID bson.ObjectId `json:"contactListID" bson:"contactListID"`
}

//type PurchaseNumbers struct {
//	PurchaseNumbers string `json:"phone_number"`
//}
//
//
//
//
//
//type PurchaseNumbersRequest struct {
//	TTSListID bson.ObjectId `json:"contactListID" bson:"contactListID"`
//	Numbers []string `json:"phone_numbers"`
//	UserID     bson.ObjectId ` bson:"userid"`
//}
//
//type TelnyxPurchaseRequest struct {
//	Numbers []PurchaseNumbers `json:"phone_numbers"`
//}
//
//
//type searchnumbers struct {
//	Data []struct {
//		BestEffort      bool `json:"best_effort"`
//		CostInformation struct {
//			Currency    string `json:"currency"`
//			MonthlyCost string `json:"monthly_cost"`
//			UpfrontCost string `json:"upfront_cost"`
//		} `json:"cost_information"`
//		Features []struct {
//			Name string `json:"name"`
//		} `json:"features"`
//		PhoneNumber       string `json:"phone_number"`
//		RecordType        string `json:"record_type"`
//		RegionInformation []struct {
//			RegionName string `json:"region_name"`
//			RegionType string `json:"region_type"`
//		} `json:"region_information"`
//		Reservable   bool        `json:"reservable"`
//		VanityFormat interface{} `json:"vanity_format"`
//	} `json:"data"`
//	Metadata struct {
//		BestEffortResults int `json:"best_effort_results"`
//		TotalResults      int `json:"total_results"`
//	} `json:"metadata"`
//	URL string `json:"url"`
//}
//

//type MyJsonName struct {
//	Data struct {
//		ConnectionID       interface{} `json:"connection_id"`
//		CreatedAt          string      `json:"created_at"`
//		CustomerReference  interface{} `json:"customer_reference"`
//		ID                 string      `json:"id"`
//		MessagingProfileID interface{} `json:"messaging_profile_id"`
//		PhoneNumbers       []struct {
//			ID                     string        `json:"id"`
//			PhoneNumber            string        `json:"phone_number"`
//			RecordType             string        `json:"record_type"`
//			RegulatoryRequirements []interface{} `json:"regulatory_requirements"`
//			RequirementsMet        bool          `json:"requirements_met"`
//			Status                 string        `json:"status"`
//		} `json:"phone_numbers"`
//		PhoneNumbersCount int64  `json:"phone_numbers_count"`
//		RecordType        string `json:"record_type"`
//		RequirementsMet   bool   `json:"requirements_met"`
//		Status            string `json:"status"`
//		UpdatedAt         string `json:"updated_at"`
//	} `json:"data"`
//}
//
//

//
//func (telnyx *NumbersCarrierGroupService) CreateMultiple(op TelnyxNumbersGroupCreateMultiple) ([]*model.TelnyxNumbersGroup,error) {
//
//	var objects []*model.TelnyxNumbersGroup
//	for i, _ := range op.NumberID {
//		var object model.TelnyxNumbersGroup
//		object.UserID = op.UserID
//		object.TTSListID = op.TTSListID
//		if err := !bson.IsObjectIdHex(op.NumberID[i]); err != false {
//			return nil,errors.New("invalid Object ID")
//		}
//		object.NumberID =bson.ObjectIdHex(op.NumberID[i])
//		contactlistid, err := telnyx.ContactListService.ContactListObjectWithId(op.TTSListID.Hex())
//		if err != nil {
//			return nil,err
//		}
//		lastcontactgroup := contactlistid.ContactGroups[(len(contactlistid.ContactGroups) - 1)]
//		checktotal, err := telnyx.ContactGroupCount(lastcontactgroup.ContactGroupId)
//		if err!=nil{
//			return nil,err
//		}
//		number, err := telnyx.NumberService.GetNumberByID(object.NumberID)
//		if err != nil {
//			return nil,err
//		}
//		if checktotal <= 10000 {
//			telnyx.PushInContactGroup(lastcontactgroup.ContactGroupId, number.PhoneNumber)
//		}
//		object.ContactGroupID = lastcontactgroup.ContactGroupId
//
//		result,err1 := telnyx.StoreTelnyxNumberGroup(&object)
//		if err1 != nil {
//			return nil,err1
//		}
//		objects = append(objects,result)
//
//	}
//	return objects,nil
//}

func (telnyx *NumbersCarrierGroupService) CreateMultiple(op TelnyxNumbersGroupCreateMultiple) ([]*model.TelnyxNumbersGroup, error) {

	var objects []*model.TelnyxNumbersGroup
	var options TelnyxNumbersGroupCreateSingle
	for i, _ := range op.NumberID {
		options.UserID = op.UserID
		options.NumberID = op.NumberID[i]
		options.ContactListID = op.ContactListID
		res, err := telnyx.CreateSingle(options)
		if err != nil {
			return nil, err
		}
		objects = append(objects, res)

	}
	return objects, nil
}

func (telnyx *NumbersCarrierGroupService) CreateSingle(op TelnyxNumbersGroupCreateSingle) (*model.TelnyxNumbersGroup, error) {
	var object model.TelnyxNumbersGroup
	object.UserID = op.UserID
	object.ContactListID = op.ContactListID
	if err := !bson.IsObjectIdHex(op.NumberID); err != false {
		return nil, errors.New("invalid Object ID")
	}
	object.NumberID = bson.ObjectIdHex(op.NumberID)
	contactlistid, err := telnyx.ContactListService.ContactListObjectWithId(op.ContactListID.Hex())
	if err != nil {
		return nil, err
	}
	lastcontactgroup := contactlistid.ContactGroups[(len(contactlistid.ContactGroups) - 1)]
	checktotal, err := telnyx.ContactGroupCount(lastcontactgroup.ContactGroupId)
	if err != nil {
		return nil, err
	}

	number, err := telnyx.NumberService.GetNumberByID(object.NumberID)
	if err != nil {
		return nil, err
	}
	checknumber, err := telnyx.CheckNumberInContactGroup(lastcontactgroup.ContactGroupId, number.PhoneNumber)
	if checknumber == true {
		object.ContactGroupID = lastcontactgroup.ContactGroupId
		res, err := telnyx.GetOneTelnyxNumberGroup(object)
		if err != nil {
			resultt, err1 := telnyx.StoreTelnyxNumberGroup(&object)
			if err1 != nil {
				return nil, err1
			}
			return resultt, nil
		}
		return res, nil
	}
	if checktotal <= 10000 {
		telnyx.PushInContactGroup(lastcontactgroup.ContactGroupId, number.PhoneNumber)
		telnyx.IncrementContactGroupNumberCount(lastcontactgroup.ContactGroupId)
		telnyx.IncrementContactListNumberCount(contactlistid.ID)
	}
	object.ContactGroupID = lastcontactgroup.ContactGroupId
	result, err1 := telnyx.StoreTelnyxNumberGroup(&object)
	if err1 != nil {
		return nil, err1
	}

	return result, nil
}

func (telnyx *NumbersCarrierGroupService) GetContactListStats(op model.GetContactListStatsRequest) (*[]model.ContactListStats, int, error) {
	//fmt.Println(userObjectId)
	// get new instance of campaign list datasource
	list, total, err := telnyx.GetAllContactList(op)
	if err != nil {
		return nil, total, err
	}

	return list, total, nil
}

func (telnyx *NumbersCarrierGroupService) GetAll(op model.GetAllTelnyxGroupRequest) (*[]model.TelnyxNumbersGroupAggr, int, error) {
	//fmt.Println(userObjectId)
	// get new instance of campaign list datasource
	list, total, err := telnyx.GetAllTelnyxNumberGroup(op)
	if err != nil {
		return nil, total, err
	}

	return list, total, nil
}

//func (telnyx *NumbersCarrierGroupService) DeleteMultiple(op TelnyxNumbersGroupDeleteMultipleRequest) error {
//	for i, _ := range op.TTSListID {
//		var object model.TelnyxNumbersGroup
//		object.UserID = op.UserID
//		if err := !bson.IsObjectIdHex(op.NumberID); err != false {
//			return errors.New("invalid Object ID")
//		}
//		cid := bson.ObjectIdHex(op.TTSListID[i])
//		object.TTSListID = cid
//		tid :=bson.ObjectIdHex(op.NumberID)
//		object.NumberID =tid
//		fmt.Println(object.UserID)
//		fmt.Println(object.TTSListID[i])
//		id, err := telnyx.NumberService.GetNumberByID(object.NumberID)
//		if err != nil {
//			return err
//		}
//		contactgroupid, err := telnyx.GetByID(&object)
//		if err != nil {
//			return err
//		}
//		var numberobject model.TelnyxContactGroup
//		numberobject.Number = id.PhoneNumber
//		numberobject.ContactGroupID = contactgroupid.ContactGroupID
//		numberobject.UserID = op.UserID
//		fmt.Println(numberobject)
//		err2 := telnyx.ContactGroupService.DeleteTelnyxContactGroup(&numberobject)
//
//		if err2 != nil {
//			return err2
//		}
//		err1 := telnyx.DeleteTelnyxNumberGroup(&object)
//		if err1 != nil {
//			return err1
//		}
//	}
//	return nil
//}

func (telnyx *NumbersCarrierGroupService) DeleteMultiple(op TelnyxNumbersGroupDeleteMultipleRequest) error {
	var options TelnyxNumbersGroupDeleteSingleRequest
	for i, _ := range op.ContactListID {
		options.ContactListID = bson.ObjectIdHex(op.ContactListID[i])
		options.UserID = op.UserID
		options.NumberID = op.NumberID
		err := telnyx.DeleteSingle(options)
		if err != nil {
			return err
		}

	}
	return nil
}

func (telnyx *NumbersCarrierGroupService) DeleteSingle(op TelnyxNumbersGroupDeleteSingleRequest) error {
	var object model.TelnyxNumbersGroup
	object.UserID = op.UserID
	object.ContactListID = op.ContactListID
	object.NumberID = bson.ObjectIdHex(op.NumberID)
	if err := !bson.IsObjectIdHex(op.NumberID); err != false {
		return errors.New("invalid Object ID")
	}
	id, err := telnyx.NumberService.GetNumberByID(object.NumberID)

	if err != nil {
		return err
	}
	object.NumberID = bson.ObjectIdHex(op.NumberID)
	contactgroupid, err := telnyx.GetByID(&object)
	if err != nil {
		return err
	}
	var numberobject model.TelnyxContactGroup
	numberobject.Number = id.PhoneNumber
	numberobject.ContactGroupID = contactgroupid.ContactGroupID
	numberobject.UserID = op.UserID

	err2 := telnyx.ContactGroupService.DeleteTelnyxContactGroup(&numberobject)
	if err2 != nil {
		return err2
	}
	telnyx.DecrementContactGroupNumberCount(numberobject.ContactGroupID)
	telnyx.DecrementContactListNumberCount(object.ContactListID)
	err1 := telnyx.DeleteTelnyxNumberGroup(&object)
	if err1 != nil {
		return err1
	}

	return nil
}

//
func (telnyx *NumbersCarrierGroupService) StoreTelnyxNumberGroup(obj *model.TelnyxNumbersGroup) (*model.TelnyxNumbersGroup, error) {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxNumbersGroupDataSource()
	defer tnd.Session.Close()
	return tnd.Store(obj)
}

func (telnyx *NumbersCarrierGroupService) DeleteTelnyxNumberGroup(obj *model.TelnyxNumbersGroup) error {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxNumbersGroupDataSource()
	defer tnd.Session.Close()
	return tnd.Delete(obj)
}

func (telnyx *NumbersCarrierGroupService) GetAllTelnyxNumberGroup(op model.GetAllTelnyxGroupRequest) (*[]model.TelnyxNumbersGroupAggr, int, error) {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxNumbersGroupDataSource()
	defer tnd.Session.Close()
	return tnd.GetAll(op)
}

func (telnyx *NumbersCarrierGroupService) DeleteNumberGroupsByContactListID(userid string, contactlistid string) error {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxNumbersGroupDataSource()
	defer tnd.Session.Close()
	user_id := bson.ObjectIdHex(userid)
	contactlistID := bson.ObjectIdHex(contactlistid)
	return tnd.DeleteNumberGroupsByContactListID(user_id, contactlistID)
}

func (telnyx *NumbersCarrierGroupService) GetOneTelnyxNumberGroup(op model.TelnyxNumbersGroup) (*model.TelnyxNumbersGroup, error) {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxNumbersGroupDataSource()
	defer tnd.Session.Close()
	return tnd.GetOneTelnyxNumberGroup(op)
}

func (telnyx *NumbersCarrierGroupService) GetAllContactList(op model.GetContactListStatsRequest) (*[]model.ContactListStats, int, error) {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxNumbersGroupDataSource()
	defer tnd.Session.Close()
	return tnd.GetContactListStats(op)
}

func (telnyx *NumbersCarrierGroupService) GetByID(obj *model.TelnyxNumbersGroup) (*model.TelnyxNumbersGroup, error) {
	// get new instance of sound file datasource
	tnd := telnyx.telnyxNumbersGroupDataSource()
	defer tnd.Session.Close()
	return tnd.GetByID(obj)
}

func (telnyx *NumbersCarrierGroupService) telnyxNumbersGroupDataSource() *datasource.TelnyxNumbersGroupDataSource {

	return &datasource.TelnyxNumbersGroupDataSource{DataSource: datasource.DataSource{Session: telnyx.Session.Copy()}}
}

//
//func (telnyx *TelnyxService) SaveTelnyxNumberCID(obj *model.TelnyxNumberCID) error {
//	// get new instance of sound file datasource
//	tnd := telnyx.telnyxDataSource()
//	defer tnd.Session.Close()
//	return tnd.SaveTelnyxCID(obj)
//}
//
//
//
//func (telnyx *TelnyxService) UpdateTelnyxNumber(obj *model.TelnyxNumber) error {
//	// get new instance of sound file datasource
//	tnd := telnyx.telnyxDataSource()
//	defer tnd.Session.Close()
//	return tnd.Update(obj)
//}
//
//
//func (telnyx *TelnyxService) GetTelnyxNumber(obj string) (model.TelnyxNumber,error) {
//	// get new instance of sound file datasource
//	tnd := telnyx.telnyxDataSource()
//	defer tnd.Session.Close()
//	return tnd.Get(obj)
//}
//
//
//func (telnyx *TelnyxService) DeleteTelnyxNumber(id string) error {
//	// get new instance of sound file datasource
//	tnd := telnyx.telnyxDataSource()
//	defer tnd.Session.Close()
//	return tnd.DeleteWithId(id)
//}
//
//func (telnyx *TelnyxService) telnyxDataSource() *datasource.TelnyxDataSource {
//	return &datasource.TelnyxDataSource{DataSource: datasource.DataSource{Session: telnyx.Session.Copy()}}
//}
//
//func (telnyx *TelnyxService) DeleteNumbers(id string) error {
//
//	client := &http.Client{}
//	url := "https://api.telnyx.com/v2/phone_numbers/" + id
//	req, err := http.NewRequest("DELETE", url, nil)
//	req.Header.Add("Content-Type", "application/json")
//	req.Header.Add("Accept", "application/json")
//	telnyx_key := "Bearer " + os.Getenv("TELNYX_API_KEY")
//	req.Header.Add("Authorization", telnyx_key)
//	fmt.Println(req)
//	if err != nil {
//		log.Fatal(err)
//	}
//	resp, err := client.Do(req)
//	if err != nil {
//		return err
//	}
//	fmt.Println("response Status:", resp.Status)
//	fmt.Println("response Headers:", resp.Header)
//	body, _ := ioutil.ReadAll(resp.Body)
//	fmt.Println("response Body:", string(body))
//	//myText := strings.Replace(id, "-", "", 1)
//	//myText = strings.TrimSpace(id)
//	//fmt.Println(myText)
//	err1 := telnyx.DeleteTelnyxNumber(id)
//	if err1 != nil {
//		fmt.Println(err1.Error())
//	}
//	return nil
//}
//
//func (telnyx *TelnyxService) PurchaseNumbers(op PurchaseNumbersRequest) (PurchaseNumbersRequest,error) {
//	client := &http.Client{}
//	box := []PurchaseNumbers{}
//	for _, v := range op.Numbers {
//		box = append(box, PurchaseNumbers{v})
//	}
//	fmt.Println(box)
//	boxx := TelnyxPurchaseRequest{}
//	boxx.Numbers = box
//	fmt.Println(boxx)
//	b, err := json.Marshal(boxx)
//	if err != nil {
//		fmt.Println(err)
//		return  op,err
//	}
//	fmt.Println(string(b))
//	reqbody := []byte(b)
//	fmt.Println(reqbody)
//	req, err := http.NewRequest("POST", "https://api.telnyx.com/v2/number_orders", bytes.NewBuffer(reqbody))
//	req.Header.Add("Content-Type", "application/json")
//	req.Header.Add("Accept", "application/json")
//	telnyx_key := "Bearer " + os.Getenv("TELNYX_API_KEY")
//	req.Header.Add("Authorization", telnyx_key)
//	fmt.Println(req)
//	if err != nil {
//		log.Fatal(err)
//	}
//	resp, err := client.Do(req)
//
//	fmt.Println("response Status:", resp.Status)
//	fmt.Println("response Headers:", resp.Header)
//	body, _ := ioutil.ReadAll(resp.Body)
//	fmt.Println("response Body:", string(body))
//	var result MyJsonName
//	json.Unmarshal(body, &result)
//	fmt.Println(result)
//	fmt.Println(result.Data.Status)
//	fmt.Println(result.Data.PhoneNumbers[0].ID)
//
//	for i, _ := range op.Numbers {
//		var number model.TelnyxNumber
//		number.PhoneID = string(result.Data.PhoneNumbers[i].ID)
//		number.PhoneNumber = result.Data.PhoneNumbers[i].PhoneNumber
//		number.PhoneStatus = result.Data.PhoneNumbers[i].Status
//		number.InboundCalls = 5
//		number.Callbacks = 5.6
//		fmt.Println(number)
//		reqbody := []byte(`{"connection_id":"1460742985854486497"}`)
//
//		url := "https://api.telnyx.com/v2/phone_numbers/" + result.Data.PhoneNumbers[i].PhoneNumber
//		fmt.Println(url)
//		req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(reqbody))
//		req.Header.Add("Content-Type", "application/json")
//		req.Header.Add("Accept", "application/json")
//		telnyx_key := "Bearer " + os.Getenv("TELNYX_API_KEY")
//		req.Header.Add("Authorization", telnyx_key)
//		fmt.Println(req)
//		if err != nil {
//			log.Fatal(err)
//		}
//		resp, err := client.Do(req)
//		defer resp.Body.Close()
//		fmt.Println("response Status:", resp.Status)
//		fmt.Println("response Headers:", resp.Header)
//		body, _ := ioutil.ReadAll(resp.Body)
//		fmt.Println("response Body:", string(body))
//
//		number.ConnectionID = "1460742985854486497"
//		number.UserID = op.UserID
//		err1 := telnyx.SaveTelnyxNumber(&number)
//		if err1 != nil {
//			fmt.Println(err1.Error())
//		}
//
//	}
//	gocron.Every(15).Second().Do(telnyx.CheckStatusTelnyxNumbers, op)
//
//	//result, err2 := telnyx.GetTelnyxNumber(string(number.PhoneNumber))
//	//if err2 != nil {
//	//	fmt.Println(err1.Error())
//	//}
//	//4
//	//var telcid model.TelnyxNumberCID
//	//telcid.TTSListID = op.TTSListID
//	//telcid.UserID = op.UserID
//	//telcid.TelnyxID = result.ID
//	//err3 := telnyx.SaveTelnyxNumberCID(&telcid)
//	//if err3 != nil {
//	//	fmt.Println(err1.Error())
//	//}
//	//
//
//	//obj,err:=telnyx.ContactListService.ContactListObjectWithId(op.TTSListID.Hex())
//	//if err != nil{
//	//	fmt.Println(err)
//	//}
//	//fmt.Println(obj)
//	// := telnyx.contactListDataSource()
//	//defer telnyxDataSource.Session.Close()
//	//def,err:=telnyxDataSource.ContactListWithId(op.TTSListID)
//	//fmt.Println(def)
//	//var contactlistobject *model.ContactList
//	//tnd := telnyx.ContactListService.contactListDataSource()
//	//contactlistobject ,err:=tnd.ContactListWithId(op.TTSListID)
//	//defer tnd.Session.Close()
//	//
//	//	fmt.Println(contactlistobject)
//	//fmt.Println("here")
//		//contactgroups := contactlistobject.ContactGroups
//		//fmt.Println(contactgroups)
//
//	return op,err
//	}
//
//
//
//	//fmt.Println(number)
//	// err1 := telnyx.SaveTelnyxNumber(&number)
//	//if err1!=nil{
//	//	fmt.Println(err1.Error())
//	//}
//	//fmt.Println(result.Data.PhoneNumbers[1])
//	//fmt.Println(result.Data.PhoneNumbers)
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//	//for key, element := range result {
//	//	fmt.Println("Key:", key, "=>", "Element:", element)
//	//}
//	//fmt.Println(result)
//	//fmt.Println(result["data"].(map[string]interface{})["phone_numbers"])
//	//fmt.Println(result["id"])
//	//json.NewDecoder(resp.Body).Decode(&result)
//	//order_id := result["data"].(map[string]interface{})["id"]
//	////order_status := result["data"].(map[string]interface{})["status"]
//	//phone_id := result["data"].(map[string]interface{})["phone_numbers"].(map[string]interface{})["id"]
//	//phone_status := result["data"].(map[string]interface{})["phone_numbers"].(map[string]interface{})["status"]
//	//phone_number := result["data"].(map[string]interface{})["phone_numbers"].(map[string]interface{})["phone_number"]
//	//fmt.Println(order_id)
//	//fmt.Println(phone_id)
//	////fmt.Println(order_status)
//	//fmt.Println(phone_status)
//	//fmt.Println(phone_number)
//
//	//fmt.Println(result["id"].(string))
//	//fmt.Println(result["phone_numbers"].(string))
//	//
//	//messageupdate := `{"connection_id": "1460742985854486497"}`
//	//var reqbodyy = []byte(messageupdate)
//	//url := "https://api.telnyx.com/v2/phone_numbers/" + phone_id
//	//reqq, err := http.NewRequest("Patch", url, bytes.NewBuffer(reqbodyy))
//	//reqq.Header.Add("Accept", "application/json")
//	//reqq.Header.Add("Authorization", telnyx_key)
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//	//respp, err := client.Do(req)
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//	//defer respp.Body.Close()
//	//var resultt map[string]interface{}
//	//json.NewDecoder(resp.Body).Decode(&resultt)
//	//
//	//fmt.Println(resultt["connection_id"].(string))
//
//	//return op,nil
//
//
//
//
//
//func (telnyx *TelnyxService) CheckStatusTelnyxNumbers(op PurchaseNumbersRequest) (error) {
//
//	client := &http.Client{}
//	for i, _ := range op.Numbers {
//		var number model.TelnyxNumber
//		//number.OrderID = string(result.Data.ID)
//		number,err := telnyx.GetTelnyxNumber(op.Numbers[i])
//		fmt.Println(number.PhoneNumber)
//		if err!=nil{
//		fmt.Println(err)
//		}
//		url := "https://api.telnyx.com/v2/phone_numbers/" + number.PhoneNumber
//		fmt.Println(url)
//		req, err := http.NewRequest("GET", url, nil)
//		req.Header.Add("Content-Type", "application/json")
//		req.Header.Add("Accept", "application/json")
//		telnyx_key := "Bearer " + os.Getenv("TELNYX_API_KEY")
//		req.Header.Add("Authorization", telnyx_key)
//		fmt.Println(req)
//		if err != nil {
//			log.Fatal(err)
//		}
//		resp, err := client.Do(req)
//		defer resp.Body.Close()
//		fmt.Println("response Status:", resp.Status)
//		fmt.Println("response Headers:", resp.Header)
//		body, _ := ioutil.ReadAll(resp.Body)
//		fmt.Println("response Body:", string(body))
//
//		var result StatusResponseTelnyx
//		json.Unmarshal(body, &result)
//
//		if (result.Data.Status == "active") {
//			gocron.Remove(telnyx.CheckStatusTelnyxNumbers)
//			gocron.Clear()
//			number.PhoneNumber= result.Data.PhoneNumber
//			number.PhoneStatus = result.Data.Status
//			fmt.Println(result.Data.Status)
//			fmt.Println(result.Data.PhoneNumber)
//			err1 := telnyx.UpdateTelnyxNumber(&number)
//			if err1 != nil {
//				fmt.Println(err1.Error())
//			}
//		}
//
//	}
//	return nil
//}

//func (telnyx *TelnyxService) ListTelnyxNumbers(objectID string, page int, limit int) (*model.TelnyxList, int, error) {
//	if !bson.IsObjectIdHex(objectID) {
//		return nil, 0, errors.New("User object id is not valid")
//	}
//
//	userObjectId := bson.ObjectIdHex(objectID)
//	//fmt.Println(userObjectId)
//	// get new instance of campaign list datasource
//	telnyxDataSource := telnyx.telnyxDataSource()
//	defer telnyxDataSource.Session.Close()
//
//	list, total, err := telnyxDataSource.TelnyxListGetList(userObjectId, page, limit)
//	if err != nil {
//		return nil, total, err
//	}
//
//	return list, total, nil
//}
