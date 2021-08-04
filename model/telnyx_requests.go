package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type TelnyxStatusResponse struct {
	Data struct {
		ID                    string        `json:"id"`
		RecordType            string        `json:"record_type"`
		PhoneNumber           string        `json:"phone_number"`
		Status                string        `json:"status"`
		Tags                  []interface{} `json:"tags"`
		ConnectionID          string        `json:"connection_id"`
		ExternalPin           interface{}   `json:"external_pin"`
		T38FaxGatewayEnabled  bool          `json:"t38_fax_gateway_enabled"`
		PurchasedAt           time.Time     `json:"purchased_at"`
		BillingGroupID        interface{}   `json:"billing_group_id"`
		EmergencyEnabled      bool          `json:"emergency_enabled"`
		CallForwardingEnabled bool          `json:"call_forwarding_enabled"`
		CnamListingEnabled    bool          `json:"cnam_listing_enabled"`
		CallRecordingEnabled  bool          `json:"call_recording_enabled"`
		MessagingProfileID    string        `json:"messaging_profile_id"`
		MessagingProfileName  string        `json:"messaging_profile_name"`
		CreatedAt             time.Time     `json:"created_at"`
		UpdatedAt             time.Time     `json:"updated_at"`
	} `json:"data"`
}

type GetAllTelnyxGroupRequest struct {
	NumberID bson.ObjectId
	UserID   bson.ObjectId
	Page     int
	Limit    int
	Search   string
}

type GetContactListStatsRequest struct {
	ContactListID bson.ObjectId
	UserID        bson.ObjectId
	Page          int
	Limit         int
	Search        string
}

type GetAllRequest struct {
	UserID bson.ObjectId
	Page   int
	Limit  int
	Search string
}

type UpdateSingleNumberRequest struct {
	ForwardingNumber string `json:"forwardingNumber"`
	NumberID         string
}

type BulkVsSearchRequest struct {
	Npa string
}

type TelnyxSearchNumbersRequest struct {
	NumberType       string
	AreaCode         string
	Locality         string
	AdministriveArea string
	Limit            string
	Npa              string
}

type BaseTelnyxPurchaseRequest struct {
	PurchaseNumbers string `json:"phone_number"`
}
type FinalTelnyxPurchaseRequest struct {
	Numbers []BaseTelnyxPurchaseRequest `json:"phone_numbers"`
}

type UpdateMultipleNumberRequest struct {
	//TTSListID bson.ObjectId `json:"contactListID" bson:"contactListID"`
	NumberID         []string      `json:"numberID"`
	UserID           bson.ObjectId `bson:"userid"`
	ForwardingNumber string        `json:"forwardingNumber"`
}

type TelnyxSearchResponse struct {
	Data []struct {
		//BestEffort      bool `json:"best_effort"`
		CostInformation struct {
			Currency    string `json:"currency"`
			MonthlyCost string `json:"monthly_cost"`
			UpfrontCost string `json:"upfront_cost"`
		} `json:"cost_information"`
		Features []struct {
			Name string `json:"name"`
		} `json:"features"`
		PhoneNumber string `json:"phone_number"`
		//RecordType        string `json:"record_type"`
		RegionInformation []struct {
			RegionName string `json:"region_name"`
			RegionType string `json:"region_type"`
		} `json:"region_information"`
		//Reservable   bool        `json:"reservable"`
		//VanityFormat interface{} `json:"vanity_format"`
	} `json:"data"`
	Metadata struct {
		//	BestEffortResults int `json:"best_effort_results"`
		TotalResults int `json:"total_results"`
	} `json:"metadata"`
	URL string `json:"url"`
}

type BulkVSSearchResponse []struct {
	TN            string `json:"TN"`
	RateCenter    string `json:"Rate Center"`
	State         string `json:"State"`
	PerMinuteRate string `json:"Per Minute Rate"`
	Mrc           string `json:"Mrc"`
	Nrc           string `json:"Nrc"`
}

type TelnyxPurchaseResponse struct {
	Data struct {
		ConnectionID       interface{} `json:"connection_id"`
		CreatedAt          string      `json:"created_at"`
		CustomerReference  interface{} `json:"customer_reference"`
		ID                 string      `json:"id"`
		MessagingProfileID interface{} `json:"messaging_profile_id"`
		PhoneNumbers       []struct {
			ID                     string        `json:"id"`
			PhoneNumber            string        `json:"phone_number"`
			RecordType             string        `json:"record_type"`
			RegulatoryRequirements []interface{} `json:"regulatory_requirements"`
			RequirementsMet        bool          `json:"requirements_met"`
			Status                 string        `json:"status"`
		} `json:"phone_numbers"`
		PhoneNumbersCount int64  `json:"phone_numbers_count"`
		RecordType        string `json:"record_type"`
		RequirementsMet   bool   `json:"requirements_met"`
		Status            string `json:"status"`
		UpdatedAt         string `json:"updated_at"`
	} `json:"data"`
}

type PurchaseNumbersRequest struct {
	Numbers []string      `json:"phoneNumbers"`
	UserID  bson.ObjectId `bson:"userid"`
}

type BulkvsPurchaseNumbersRequest struct {
	Numbers []string      `json:"phoneNumbers"`
	UserID  bson.ObjectId `bson:"userid"`
}

type DeleteNumbersRequest struct {
	NumberID []string      `json:"numberID"`
	UserID   bson.ObjectId `bson:"userid"`
}
