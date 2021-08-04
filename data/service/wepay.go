package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/model"
)

type WePay struct {
}

type GetAccountRequest struct {
	AccountId int64 `json:"account_id"`
}

type AuthorizeRequest struct {
	ClientId     int64  `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	Code         string `json:"code"`
}

type AuthorizeResponse struct {
	AccessToken string `json:"access_token"`
}

type CreateAccountRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ReferenceId string `json:"reference_id"`
}

type UpdateURIRequest struct {
	AccountId   int64  `json:"account_id"`
	RedirectURI string `json:"redirect_uri"`
	Mode        string `json:"mode"`
}

type CheckoutCreateRequest struct {
	AccountId        int64                 `json:"account_id"`
	Amount           float32               `json:"amount"`
	Type             string                `json:"type"`
	Currency         string                `json:"currency"`
	ShortDescription string                `json:"short_description"`
	HostedCheckout   HostedCheckoutRequest `json:"hosted_checkout"`
	CallbackUri      string                `json:"callback_uri"`
}

type GetNotificationPreferenceRequest struct {
	NotificationPreferenceId string `json:"notification_preference_id"`
}

type NotificationPreferenceResponse struct {
	NotificationPreferenceId string `json:"notification_preference_id"`
	Type                     string `json:"type"`
	AppId                    int64  `json:"app_id"`
	Topic                    string `json:"topic"`
	CallbackUri              string `json:"callback_uri"`
	State                    string `json:"state"`
}

type CreateNotificationPreferenceRequest struct {
	Type        string `json:"type"`
	Topic       string `json:"topic"`
	CallbackUri string `json:"callback_uri"`
}

type HostedCheckoutRequest struct {
	RedirectURI string `json:"redirect_uri"`
	Mode        string `json:"mode"`
}

type UpdateURIResponse struct {
	AccountId int64  `json:"account_id"`
	URI       string `json:"uri"`
}

type WePayAccountInfo struct {
	AccountId   int64  `json:"account_id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	State       string `json:"state"`
	ReferenceId string `json:"reference_id"`
}

type CreateAccountOptions struct {
	AccessToken string
	Name        string
	Description string
	ReferenceId string
}

func NewWePayService() *WePay {
	return &WePay{}
}

func (s *WePay) GetAccount(accountToken string, accountId int64) (*WePayAccountInfo, error) {
	wePayApi := os.Getenv("WEPAY_API")
	api := fmt.Sprintf("%s/account", wePayApi)

	body := GetAccountRequest{AccountId: accountId}
	info, _ := json.Marshal(body)
	hc := http.Client{}
	req, err := http.NewRequest("POST", api, bytes.NewBuffer(info))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accountToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var responseObject WePayAccountInfo
		json.Unmarshal(responseData, &responseObject)
		return &responseObject, nil
	}

	return nil, errors.New("Error fetching account")
}

func (s *WePay) AuthorizeToken(redirectUrl string, code string) (string, error) {
	wePayApi := os.Getenv("WEPAY_API")
	clientId := os.Getenv("WEPAY_CLIENT_ID")
	clientSecret := os.Getenv("WEPAY_CLIENT_SECRET")
	api := fmt.Sprintf("%s/oauth2/token", wePayApi)

	var clientIdInt int
	clientIdInt, _ = strconv.Atoi(clientId)

	body := AuthorizeRequest{ClientId: int64(clientIdInt), ClientSecret: clientSecret, RedirectURI: redirectUrl, Code: code}
	info, _ := json.Marshal(body)
	hc := http.Client{}
	req, err := http.NewRequest("POST", api, bytes.NewBuffer(info))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		var responseObject AuthorizeResponse
		json.Unmarshal(responseData, &responseObject)
		return responseObject.AccessToken, nil
	}

	return "", errors.New("Error Authorization")
}

func (s *WePay) CreateAccount(options *CreateAccountOptions) (*WePayAccountInfo, error) {
	wePayApi := os.Getenv("WEPAY_API")
	api := fmt.Sprintf("%s/account/create", wePayApi)

	body := CreateAccountRequest{Name: options.Name, Description: options.Description, ReferenceId: options.ReferenceId}
	info, _ := json.Marshal(body)
	hc := http.Client{}
	req, err := http.NewRequest("POST", api, bytes.NewBuffer(info))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", options.AccessToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var responseObject WePayAccountInfo
		json.Unmarshal(responseData, &responseObject)
		return &responseObject, nil
	}

	return nil, errors.New("Error fetching account")
}

func (s *WePay) UpdateAccountURI(accessToken string, accountId int64, redirectURI string) (string, error) {
	wePayApi := os.Getenv("WEPAY_API")
	api := fmt.Sprintf("%s/account/get_update_uri", wePayApi)

	body := UpdateURIRequest{AccountId: accountId, RedirectURI: redirectURI, Mode: "iframe"}
	info, _ := json.Marshal(body)
	hc := http.Client{}
	req, err := http.NewRequest("POST", api, bytes.NewBuffer(info))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		var responseObject UpdateURIResponse
		json.Unmarshal(responseData, &responseObject)
		return responseObject.URI, nil
	}

	return "", errors.New("Error fetching account")
}

func (s *WePay) CheckoutCreate(accessToken string, accountId int64, redirectURI string, amount float32, callbackuri string) (*model.WepayCheckout, error) {
	wePayApi := os.Getenv("WEPAY_API")

	api := fmt.Sprintf("%s/checkout/create", wePayApi)

	body := CheckoutCreateRequest{
		AccountId:        accountId,
		Amount:           amount,
		Type:             "personal",
		Currency:         "USD",
		ShortDescription: "Balance add for platform",
		HostedCheckout:   HostedCheckoutRequest{RedirectURI: redirectURI, Mode: "iframe"},
		CallbackUri:      callbackuri,
	}

	info, _ := json.Marshal(body)

	logrus.Debugln("Checkout Info : ", string(info))

	hc := http.Client{}
	req, err := http.NewRequest("POST", api, bytes.NewBuffer(info))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var responseObject model.WepayCheckout
		json.Unmarshal(responseData, &responseObject)
		return &responseObject, nil
	}

	return nil, errors.New("Error creating checkout")
}

func (s *WePay) GetNotificationPreference(accessToken string, accountId int64, notfId string) (*NotificationPreferenceResponse, error) {
	wePayApi := os.Getenv("WEPAY_API")

	api := fmt.Sprintf("%s/notification_preference", wePayApi)

	body := GetNotificationPreferenceRequest{
		NotificationPreferenceId: notfId,
	}

	info, _ := json.Marshal(body)
	hc := http.Client{}
	req, err := http.NewRequest("POST", api, bytes.NewBuffer(info))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var responseObject NotificationPreferenceResponse
		json.Unmarshal(responseData, &responseObject)
		return &responseObject, nil
	}

	return nil, errors.New("Error fetching notification preference")
}

func (s *WePay) CreateNotificationPreference(accessToken string, accountId int64, callbackurl string) (*NotificationPreferenceResponse, error) {
	wePayApi := os.Getenv("WEPAY_API")

	api := fmt.Sprintf("%s/notification_preference/create", wePayApi)

	body := CreateNotificationPreferenceRequest{
		Type:        "ipn",
		Topic:       "payment.*",
		CallbackUri: callbackurl,
	}

	info, _ := json.Marshal(body)
	hc := http.Client{}
	req, err := http.NewRequest("POST", api, bytes.NewBuffer(info))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var responseObject NotificationPreferenceResponse
		json.Unmarshal(responseData, &responseObject)
		return &responseObject, nil
	}

	return nil, errors.New("Error creating notification preference")
}
