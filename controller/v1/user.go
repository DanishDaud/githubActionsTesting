package controllerv1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/auth"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/network/response"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlconstants"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlmessages"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

type UserController struct {
	Controller
	//	UserRoleService      *service.UserRoleService
	ContactListService   *service.ContactListService
	UserService          *service.UserService
	CdrService           *service.CdrService
	WePayCheckoutService *service.WepayCheckoutService
	EmailService         *service.Email
	TokenService         *service.Token
	S3Service            *service.S3Service
	FileService          *service.FileService
	BillingManager       *service.BillingManager
	PaymentService       *service.PaymentService
	BalanceService       *service.BalanceService
	RedisService         *service.RedisService
	EventService		*service.EventsService
}

type BillingDetailResponse struct {
	Contract  *model.Contract   `json:"contract"`
	CallsInfo CallsInfoResponse `json:"callsInfo"`
}

type CallsInfoResponse struct {
	Rate        float32 `json:"rate"`
	BillingType int     `json:"billingType"`
}
type (
	NumberInfo struct {
		Number string   `json:"number" validate:"required"`
		List   []string `json:"list"`
	}
)
type APITokenResponse struct {
	Token string `json:"token"`
}

// data structure to get credentials from json object
type (
	Credential struct {
		Username string `json:"username" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	TelephonicCredentials struct {
		Id   int32 `json:"id" validate:"required"`
		Code int32 `json:"code" validate:"required"`
	}
	UserAccount struct {
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		Email       string `json:"email"`
		Address     string `json:"address"`
		City        string `json:"city"`
		State       string `json:"state"`
		Country     string `json:"country"`
		ZipCode     string `json:"zipCode"`
		Phone       string `json:"phone"`
		CompanyName string `json:"companyName"`
		NewPassword string `json:"newPassword"`
	}

	UpdateUserAccount struct {
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		Address     string `json:"address"`
		City        string `json:"city"`
		State       string `json:"state"`
		Country     string `json:"country"`
		ZipCode     string `json:"zipCode"`
		Phone       string `json:"phone"`
		CompanyName string `json:"companyName"`
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}

	currentFundsResponse struct {
		FundsAvailable float32 `json:"fundsAvailable" validate:"required"`
		CallsLeft      int64   `json:"callsLeft" validate:"callsLeft"`
	}

	SupportMessage struct {
		FirstName string `json:"firstName" validate:"required"`
		LastName  string `json:"lastName" validate:"required"`
		Email     string `json:"email" validate:"required"`
		Phone     string `json:"phone"`
		Message   string `json:"message" validate:"required"`
	}

	CheckoutPaymentRequest struct {
		RedirectURI string  `json:"redirect_uri" validate:"required"`
		Amount      float32 `json:"amount" validate:"required"`
	}

	CheckoutPaymentResponse struct {
		CheckoutURI string `json:"checkout_uri"`
	}

	StripeKeyResponse struct {
		Key string `json:"key"`
	}

	StripeChargeRequest struct {
		Token  string `json:"token" validate:"required"`
		Amount int64  `json:"amount" validate:"required"`
	}

	UpdateUserResponse struct {
		FirstName    string              `json:"firstName" bson:"firstName" validate:"required"`
		LastName     string              `json:"lastName" bson:"lastName" validate:"required"`
		CompanyName  string              `json:"companyName" bson:"companyName" validate:"required"`
		Phone        string              `json:"phone" bson:"phone" validate:"required"`
		Restrictions *model.Restrictions `json:"restrictions" bson:"restrictions"`
	}

	TotalCost struct {
		TotalCost float32 `json:"total_cost"`
	}
)

// CreateUser creates a  ̰new user resource
func (uc *UserController) CreateUser(c echo.Context) error {

	// Stub an user to be populated from the body
	u := UserAccount{}
	createruser := model.User{}
	if err := c.Bind(&u); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserRequestFormatIncorrect)
	}

	// validate input request body
	if err := c.Validate(u); err != nil {
		logrus.Infoln(err)
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	// so there would be no error in this case

	parentObject, err := uc.UserService.UserWithType(1)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}
	//	u.Role = *roleObject
	u.Email = strings.ToLower(u.Email)
	temp, err := uc.UserService.UserWithEmailAddress(u.Email)
	if err == nil {
		// user already exists
		if temp.IsVerified || temp.ParentID != parentObject.ID {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserWithEmailAlreadyExists)
		}

		// return user but without jwt
		// Marshal provided interface into JSON structure
		uj, _ := json.Marshal(temp)

		return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "User is already present", "Please resend account verification email", uj)
	}

	parentUser, err := uc.UserService.UserObject(parentObject.ID.Hex())
	if err != nil {
		// parent not found return error
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.InvalidParentId)
	}

	createruser.ID = bson.NewObjectId()
	createruser.Email = u.Email
	createruser.CompanyName = u.CompanyName
	createruser.Password = u.NewPassword
	createruser.Address = u.Address
	createruser.City = u.City
	createruser.Country = u.Country
	createruser.State = u.State
	createruser.Phone = u.Phone
	createruser.ZipCode = u.ZipCode
	createruser.FirstName = u.FirstName
	createruser.LastName = u.LastName
	createruser.ParentID = parentUser.ID
	createruser.IsVerified = false
	createruser.Role = model.Client
	createruser.Account.BlockLandline = true
	createruser.Account.Balance = 0
	createruser.Account.BillingType = cmlconstants.BillingTypePerContact
	createruser.Defaults.UseDefaultRates = true
	createruser.Account.HomePageUrl = parentUser.Account.HomePageUrl
	createruser.Customize.LogoPath = ""
	createruser.Customize.PrimaryColor = ""
	createruser.Customize.SecondaryColor = ""
	createruser.TimeSetting.TimeZone = "US/Pacific"

	// create a new user in database
	if err := uc.UserService.CreateNewUser(&createruser); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserCreationFailed)
	}

	// send verification email
	if err := uc.sendVerifyEmail(&createruser, parentUser); err != nil {
		// send password reset
		logrus.Debugln("account confirm email send failed email : ", u.Email)
	}

	// send new user notification
	uc.sendNewUserEmail(&createruser, parentUser)

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusCreated, "Please verify you account", "Account verification email has been send to you email address")
}

func (uc *UserController) GetUser(c echo.Context) error {
	domain := c.QueryParam("domain")
	token := c.QueryParam("token")
	var ruser *model.User

	if domain != "" {
		user, err := uc.UserService.GetDomainOwnerForUser(domain)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.DomainOwnerDoesNotExist, err.Error())
		}

		user.Password = "*****************************"
		ruser = user

	} else if token != "" {
		user, err := uc.UserService.UserByAPIToken(token)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.DomainOwnerDoesNotExist, err.Error())
		}

		user.Password = "*****************************"
		ruser = user

	} else {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, "Please Provide Domain", "Domain is compulsory field")
	}

	// Marshal provided interface into JSON structure
	userData, e := json.Marshal(ruser)

	if e != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Owner Found", "Owner Found", userData)
}

func (uc *UserController) CustomerLoginInfo(c echo.Context) error {

	// Stub an credentials to be populated from the body
	credentials := Credential{}

	// convert json to struct
	if err := c.Bind(&credentials); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.RequestFormatIncorrect)
	}

	// validate input request body
	if err := c.Validate(credentials); err != nil {
		logrus.Infoln(err)
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	credentials.Username = strings.ToLower(credentials.Username)
	userInfo, err := uc.UserService.UserObjectWithCredentials(credentials.Username, credentials.Password)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserOrPasswordDoesNotExist)
	}

	userInfo.Password = "*****************************"

	// if account is locked
	if userInfo.IsLocked {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserAccountLocked)
	}
	if userInfo.Role == model.Super {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UnauthorizedForAction)
	}
	// Marshal provided interface into JSON structure
	userData, _ := json.Marshal(userInfo)

	if userInfo.IsVerified {
		// generate jwt token
		AssignJwtToken(c, *userInfo)
		return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
	}

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "User verification is required", "Please resend account verification email", userData)
}

func (uc *UserController) TelephonicLogin(c echo.Context) error {

	// Stub an credentials to be populated from the body
	credentials := TelephonicCredentials{}

	// convert json to struct
	if err := c.Bind(&credentials); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.RequestFormatIncorrect)
	}

	// validate input request body
	if err := c.Validate(credentials); err != nil {
		logrus.Infoln(err)
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	userInfo, err := uc.UserService.UserObjectWithTelephonicCredentials(credentials.Id, credentials.Code)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	userInfo.Password = "*****************************"

	// Marshal provided interface into JSON structure
	userData, _ := json.Marshal(userInfo)

	// generate jwt token
	if err := AssignJwtToken(c, *userInfo); err != nil {
		//TODO: Log this error properly
		logrus.Infoln("token not generated")
	}

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
}

func (uc *UserController) GetUserInfo(c echo.Context) error {
	cc := c.(model.AppContext)

	// get user object
	user, er1 := uc.UserService.UserObject(cc.ID.Hex())
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	user.Password = "***********************"

	// Marshal provided interface into JSON structure
	userData, _ := json.Marshal(user)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
}
func (uc *UserController) GetUserStatsInfo(c echo.Context) error {
	cc := c.(model.AppContext)

	count, err := uc.UserService.ActiveCampaign(cc.ID.Hex())
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	amount, err := uc.UserService.TotalCost(cc.ID.Hex())
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}
	userID := cc.ID.Hex()
	stats := uc.EventService.GetUserStats(userID)
	userstats := model.UserStats{count, stats.AnswerCount + stats.RingingCount, stats.TransferCount, amount}
	//key := userID + "_live"
	//It := uc.RedisService.RedisClient.Get(key)
	//logrus.Debugln("live  value in string", It)
	//value := It.Val()
	//state, _ := strconv.Atoi(value)
	//
	//if state < 1 {
	//	state = 0
	//}
	//userstats := model.UserStats{count, state, 0, amount}

	// Marshal provided interface into JSON structure
	userData, _ := json.Marshal(userstats)

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
}
//func (uc *UserController) GetUserStatsInfo(c echo.Context) error {
//	cc := c.(model.AppContext)
//
//	count, err := uc.UserService.ActiveCampaign(cc.ID.Hex())
//	if err != nil {
//		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
//	}
//
//	amount, err := uc.UserService.TotalCost(cc.ID.Hex())
//	if err != nil {
//		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
//	}
//	userID := cc.ID.Hex()
//
//	stats := uc.EventService.GetUserStats(userID)
//	userstats := model.UserStats{count, stats.AnswerCount + stats.RingingCount, stats.TransferCount, amount}
//	//key := userID + "_live"
//	//It := uc.RedisService.RedisClient.Get(key)
//	//logrus.Debugln("live  value in string", It)
//	//value := It.Val()
//	//state, _ := strconv.Atoi(value)
//	//
//	//if state < 1 {
//	//	state = 0
//	//}
//	//userstats := model.UserStats{count, state, 0, amount}
//
//	// Marshal provided interface into JSON structure
//	userData, _ := json.Marshal(userstats)
//	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
//}

func (uc *UserController) GetUserCallReports(c echo.Context) error {
	cc := c.(model.AppContext)
	cid := c.QueryParam("id")
	from := c.QueryParam("from")
	to := c.QueryParam("to")

	if from == "" || to == "" {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "please provide valid from and to date")
	}

	fromDate, err1 := time.Parse("2006-01-02", from)
	toDate, err2 := time.Parse("2006-01-02", to)
	toDate = toDate.Add(23 * time.Hour)
	if err1 != nil || err2 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "please provide valid from and to date")
	}

	var campIdObj bson.ObjectId
	if cid != "" && bson.IsObjectIdHex(cid) {
		campIdObj = bson.ObjectIdHex(cid)
	}
	alegstats, err := uc.CdrService.CallStatsALeg(model.UserStasPayload{cc.ID, campIdObj, fromDate, toDate})
	if err != nil {

		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	blegstats, listeningStats, err := uc.CdrService.CallStatsBLeg(model.UserStasPayload{cc.ID, campIdObj, fromDate, toDate})
	if err != nil {

		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	callreportsstats := model.CallReport{AlegInformation: *alegstats,
		BlegInforamtion:       *blegstats,
		CallLengthInformation: *listeningStats}

	logrus.Infoln(callreportsstats)
	// Marshal provided interface into JSON structure
	userData, _ := json.Marshal(callreportsstats)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
}

func (uc *UserController) GetUserCallLoc(c echo.Context) error {
	cc := c.(model.AppContext)

	logrus.Infoln(cc.ID.Hex())
	//todo get user  callCount
	stats, err := uc.CdrService.UserCallLoc(cc.ID.Hex())
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	// Marshal provided interface into JSON structure
	userData, _ := json.Marshal(stats)

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
}

func (uc *UserController) Cost(c echo.Context) error {

	cc := c.(model.AppContext)
	from := c.QueryParam("from")
	to := c.QueryParam("to")

	if from == "" || to == "" {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "please provide valid from and to date")
	}

	fromDate, err1 := time.Parse("2006-01-02", from)
	toDate, err2 := time.Parse("2006-01-02", to)
	if err1 != nil || err2 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "please provide valid from and to date")
	}
	//userid:="5f9b399c16339d6a247d0e0a"
	//id:=bson.ObjectIdHex(userid)
	cost, err := uc.CdrService.CallCost(model.GetCallCostPayload{cc.ID, fromDate, toDate})
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	totalcost := TotalCost{
		TotalCost: cost,
	}

	data, err := json.Marshal(totalcost)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Success", "Success", data)

}

func (uc *UserController) GetUserCallCount(c echo.Context) error {
	cc := c.(model.AppContext)

	logrus.Infoln(cc.ID.Hex())

	//todo get user  callCount
	stats, err := uc.CdrService.UserCallMinStats(cc.ID.Hex())
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	// Marshal provided interface into JSON structure
	userData, _ := json.Marshal(stats)

	// Marshal provided interface into JSON structure

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
}

func (uc *UserController) GetUserCallCosts(c echo.Context) error {
	cc := c.(model.AppContext)
	stats, err := uc.CdrService.UserCostStats(cc.ID.Hex())
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	// Marshal provided interface into JSON structure
	userData, _ := json.Marshal(stats)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
}

func (uc *UserController) GetUserAlegCall(c echo.Context) error {
	cc := c.(model.AppContext)
	//todo get user  callCount
	stats, err := uc.CdrService.AlegStats(cc.ID.Hex())
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	// Marshal provided interface into JSON structure
	userData, _ := json.Marshal(stats)
	logrus.Infoln(cc.ID.Hex())

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
}
func (uc *UserController) GetUserBlegCall(c echo.Context) error {
	cc := c.(model.AppContext)

	logrus.Infoln(cc.ID.Hex())
	stats, err := uc.CdrService.BlegStats(cc.ID.Hex())
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	// Marshal provided interface into JSON structure
	userData, _ := json.Marshal(stats)
	logrus.Infoln(cc.ID.Hex())
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
}

func (uc *UserController) GetUserCallDispositions(c echo.Context) error {
	cc := c.(model.AppContext)

	logrus.Infoln(cc.ID.Hex())
	stats, err := uc.CdrService.UserDispositionStats(cc.ID.Hex())
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	// Marshal provided interface into JSON structure
	userData, _ := json.Marshal(stats)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
}

func (uc *UserController) UpdateUserInfo(c echo.Context) error {
	// Grab user id for which file is being uploaded
	userID := c.Param("id")

	// Stub an user to be populated from the body
	u := UpdateUserResponse{}

	if err := c.Bind(&u); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserRequestFormatIncorrect)
	}

	// validate input request body
	if err := c.Validate(u); err != nil {
		logrus.Infoln(err)
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	// get user object
	user, er1 := uc.UserService.UserObject(userID)
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	// update first name, last name, company name, phone number
	user.FirstName = u.FirstName
	user.LastName = u.LastName
	user.CompanyName = u.CompanyName
	user.Phone = u.Phone
	//user.Restrictions = u.Restrictions

	if err := uc.UserService.SaveUser(user); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	// Marshal provided interface into JSON structure
	userData, _ := json.Marshal(user)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
}

func (uc *UserController) UpdateTimeZone(c echo.Context) error {
	// Grab user id for which file is being uploaded

	cc := c.(model.AppContext)
	// Stub an user to be populated from the body
	u := model.TimeSetting{}

	if err := c.Bind(&u); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserRequestFormatIncorrect)
	}

	// validate input request body
	if err := c.Validate(u); err != nil {
		logrus.Infoln(err)
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	// get user object
	user, er1 := uc.UserService.UserObject(cc.ID.Hex())
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}
	_, err := time.LoadLocation(u.TimeZone)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())

	}
	user.TimeSetting.TimeZone = u.TimeZone

	if err := uc.UserService.SaveUser(user); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	// Marshal provided interface into JSON structure
	userData, _ := json.Marshal(user)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
}

func (uc *UserController) UpdateUserAccount(c echo.Context) error {
	cc := c.(model.AppContext)

	// verify input password object
	// Stub an user to be populated from the body
	info := UpdateUserAccount{}

	if err := c.Bind(&info); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	// validate input request body
	if err := c.Validate(info); err != nil {
		logrus.Infoln(err)
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	// get user object
	user, er1 := uc.UserService.UserObject(cc.ID.Hex())
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	if info.OldPassword != "" && info.NewPassword != "" {
		// check if previous password matches
		if user.Password != cmlutils.GetMD5Hash(info.OldPassword) {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserOldPasswordMismatch)
		}

		// update new password
		user.Password = cmlutils.GetMD5Hash(info.NewPassword)
	}

	user.FirstName = info.FirstName
	user.LastName = info.LastName
	user.Address = info.Address
	user.City = info.City
	user.State = info.State
	user.Country = info.Country
	user.ZipCode = info.ZipCode
	user.Phone = info.Phone

	if err := uc.UserService.SaveUser(user); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	// Marshal provided interface into JSON structure
	userData, _ := json.Marshal(user)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", userData)
}

// this method finds number from dnc
func (uc *UserController) FindNumberFromDNC(c echo.Context) error {

	userId := c.Param("id")
	number := c.QueryParam("number")

	url := fmt.Sprintf(cmlutils.VOIPAPIPath()+"user/%s/dnc?number=%s", userId, number)

	clnt := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, url, c.Request().Body)
	req.Header.Add("Content-type", "application/json")

	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	resp, err := clnt.Do(req)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	defer resp.Body.Close()

	body, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	return httpresponse.CreateRawResponse(&c, resp.StatusCode, body)
}

// this method deletes number from dnc
func (uc *UserController) DeleteNumberFromDNC(c echo.Context) error {

	userId := c.Param("id")
	number := c.QueryParam("number")

	url := fmt.Sprintf(cmlutils.VOIPAPIPath()+"user/%s/dnc?number=%s", userId, number)

	clnt := &http.Client{}

	req, err := http.NewRequest(http.MethodDelete, url, c.Request().Body)
	req.Header.Add("Content-type", "application/json")

	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	resp, err := clnt.Do(req)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	defer resp.Body.Close()

	body, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	return httpresponse.CreateRawResponse(&c, resp.StatusCode, body)
}

func (uc *UserController) AddDncNumber(c echo.Context) error {
	userId := c.Param("id")
	url := fmt.Sprintf(cmlutils.VOIPAPIPath()+"user/%s/adddnc", userId)
	clnt := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, url, c.Request().Body)
	req.Header.Add("Content-type", "application/json")

	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	resp, err := clnt.Do(req)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	defer resp.Body.Close()

	body, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	return httpresponse.CreateRawResponse(&c, resp.StatusCode, body)
}
func (uc *UserController) DNCNumberDelete(c echo.Context) error {

	// Grab user id for which file is being uploaded
	cc := c.(model.AppContext)
	contactListInfoId := c.Param("cid")
	number := c.QueryParam("number")
	logrus.Infoln(number)
	user, er1 := uc.UserService.UserObject(cc.ID.Hex())
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	contactListObject, er2 := uc.ContactListService.ContactListObjectWithId(contactListInfoId)
	if er2 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListDoesNotExist)
	}
	if user.ID.Hex() != contactListObject.UserID.Hex() {
		// given user is not authorized to make this call
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.ContactListUnauthorizedForAction)
	}

	if contactListObject.Type == 3 && number != "" {
		if err := uc.ContactListService.DeleteDNCNumber(contactListObject, number); err != nil {
			logrus.Debugln("Contact list delete failed => Contact list Id : " + contactListObject.ID.Hex())
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListDeleteOperationFailed)
		}
	} else {

	}
	// update total numbers uploaded by user
	userDetail, err := uc.UserService.UserDetailObjectWithUser(user.ID.Hex())
	if err == nil {
		userDetail.TotalNumbersUploaded = userDetail.TotalNumbersUploaded - contactListObject.NumberCount
		uc.UserService.SaveUserDetail(userDetail)
	}

	// update user object in db
	if err := uc.UserService.SaveUser(user); err != nil {
		logrus.Debugln("User save failed while deleting contact list => User id : " + user.ID.Hex() + " => Contact List : " + contactListObject.ID.Hex())
	}

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Contact Number deleted successfully", "Contact Number Deleted successfully")
}
func (uc *UserController) GetDNCNumber(c echo.Context) error {

	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	number := c.QueryParam("number")
	pageNumber, errPage := strconv.Atoi(page)
	limitNumber, errLimit := strconv.Atoi(limit)

	if errPage != nil || errLimit != nil {
		pageNumber = 1
		limitNumber = 0
	}
	cc := c.(model.AppContext)
	logrus.Infoln(cc)
	logrus.Debugln("DNC Number : " + number)

	// validate if number is a valid US number
	if !cmlutils.IsUSNumber(number) {
		// is not valid US number
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Number is not a valid US number")
	}

	// extract 10 digits from US phone number (after simplification)
	number = cmlutils.ExtractNumberFromUSNumber(cmlutils.SimplifyPhoneNumber(number))

	// get user object
	user, er1 := uc.UserService.UserObject(cc.ID.Hex())
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	list, count := uc.ContactListService.HasNumberDNC(user.ID.Hex(), number, limitNumber, pageNumber)
	listlen := len(list)
	if listlen == 0 {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Number Not Found")
	}
	logrus.Infoln(list)

	pageInfo := model.Page{}
	pageInfo.Page = pageNumber
	pageInfo.Limit = limitNumber
	pageInfo.TotalCount = count

	pageData, _ := json.Marshal(pageInfo)
	data, _ := json.Marshal(list)
	return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusOK, "Success", "Number Found", data, pageData)
}
func (uc *UserController) AddDncBulk(c echo.Context) error {
	userId := c.Param("id")
	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, cmlmessages.ContactListSelectFile)
	}

	// save incoming file
	dest, err := uc.ContactListService.SaveMultipartFile(file)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, err.Error())
	}

	defer uc.FileService.DeleteFile(dest)

	file2, err := os.Open(dest)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, err.Error())
	}
	defer file2.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file2.Name()))
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, err.Error())
	}

	io.Copy(part, file2)
	writer.Close()

	host := fmt.Sprintf(cmlutils.VOIPAPIPath()+"user/%s/adddncbulk", userId)
	r, err := http.NewRequest("POST", host, body)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, err.Error())
	}
	r.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	defer resp.Body.Close()

	bodyr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	return httpresponse.CreateRawResponse(&c, resp.StatusCode, bodyr)
}

// this method returns funds availability detail number from dnc
func (uc *UserController) CurrentFunds(c echo.Context) error {
	cc := c.(model.AppContext)

	user, er1 := uc.UserService.UserObject(cc.ID.Hex())
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	billingParent, er1 := uc.UserService.UserObject(user.ParentID.Hex())
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	totalCalls := user.Account.Balance / pricePerCall(user, billingParent)

	cfr := currentFundsResponse{
		FundsAvailable: user.Account.Balance,
		CallsLeft:      int64(totalCalls),
	}

	//Marshal provided interface into JSON structure
	data, _ := json.Marshal(cfr)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", data)
}

func pricePerCall(user *model.User, billingParent *model.User) float32 {
	billingRate := user.Account.BillingRateCall
	if user.Defaults.UseDefaultRates {
		// user parent's set rates
		billingRate = billingParent.Defaults.CallRate
	}

	factor := billingRateFactor(user.Account.BillingType)
	return billingRate * factor
}

func billingRateFactor(billingType int8) float32 {
	switch billingType {
	case cmlconstants.BillingTypePerContact:
		return 1
	case cmlconstants.BillingTypePer6Secs:
		return 10
	case cmlconstants.BillingTypePer30Secs:
		return 2
	case cmlconstants.BillingTypePerMinute:
		return 1
	default:
		return 1
	}
}

func (uc *UserController) BillingParent(c echo.Context) error {
	/*	userId := c.Param("id")
		user, er1 := uc.UserService.UserObject(userId)
		if er1 != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
		}

		billingParent, er1 := uc.UserService.UserObject(user.BillingParentID.Hex())
		if er1 != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
		}

		billingParent.Password = "******************"

		// Marshal provided interface into JSON structure
		data, _ := json.Marshal(billingParent)
		return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", data)*/
	return nil
}

func (uc *UserController) WePayCheckout(c echo.Context) error {
	//userId := c.Param("id")

	return httpresponse.CreateBadResponse(&c, http.StatusUnauthorized, cmlmessages.PleaseTryAgain, "Please provide valid information")

	//body := CheckoutPaymentRequest{}
	//if err := c.Bind(&body); err != nil {
	//	return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Please provide valid information")
	//}
	//
	//// validate input values
	//if err := c.Validate(body); err != nil {
	//	return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Please provide all required fields")
	//}
	//
	//user, er1 := uc.UserService.UserObject(userId)
	//if er1 != nil {
	//	return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	//}
	//
	//billingParent, er1 := uc.UserService.UserObject(user.BillingParentID.Hex())
	//if er1 != nil {
	//	return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	//}
	//
	//// check if there is any negative balance before
	//_, negAmount, err :=  uc.BalanceService.GetNegBalance(user.ID)
	//if err != nil {
	//	return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	//}
	//
	//if negAmount > float32(body.Amount) {
	//	// charge more than negative balance
	//	return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, fmt.Sprintf("Please charge more than $%f", negAmount))
	//}
	//
	//wepayService := service.NewWePayService()
	//
	//host := c.Request().Host
	//scheme := "https"
	//callback := scheme + "://" + host + "/v1/wepay/payments"
	//logrus.Debugln("*** Callback => ", callback)
	//
	//checkout, err := wepayService.CheckoutCreate(
	//	billingParent.Payments.WePayInfo.AccessToken,
	//	billingParent.Payments.WePayInfo.AccountId,
	//	body.RedirectURI,
	//	body.Amount,
	//	callback,
	//)
	//if err != nil {
	//	return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Unable to create checkout")
	//}
	//
	//// save checkout info
	//checkout.UserID = user.ID
	//checkout.Consumed = false
	//if err := uc.WePayCheckoutService.Save(checkout); err != nil {
	//	return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Unable to save checkout")
	//}
	//
	//info := CheckoutPaymentResponse{checkout.HostedCheckout.CheckoutURI}
	//
	//// Marshal provided interface into JSON structure
	//data, _ := json.Marshal(info)
	//return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", data)
}

func (uc *UserController) StripeKey(c echo.Context) error {
	/*userId := c.Param("id")
		user, er1 := uc.UserService.UserObject(userId)
		if er1 != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
		}

		billingParent, er1 := uc.UserService.UserObject(user.BillingParentID.Hex())
		if er1 != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
		}

		if billingParent.Payments.Method != cmlconstants.PaymentMethodStripe {
			return httpresponse.CreateBadResponse(&c, http.StatusUnauthorized, cmlmessages.PleaseTryAgain, cmlmessages.UnauthorizedForAction)
		}

		info := StripeKeyResponse{Key: billingParent.Payments.StripePublishableKey}
		// Marshal provided interface into JSON structure
		data, _ := json.Marshal(info)
		return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Success", "Success", data)
	}

	func (uc *UserController) StripeCharge(c echo.Context) error {
		userId := c.Param("id")

		body := StripeChargeRequest{}

		if err := c.Bind(&body); err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Please provide valid information")
		}

		// validate input values
		if err := c.Validate(body); err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Please provide all required fields")
		}
		user, er1 := uc.UserService.UserObject(userId)
		if er1 != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
		}

		billingParent, er1 := uc.UserService.UserObject(user.BillingParentID.Hex())
		if er1 != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
		}

		if billingParent.Payments.Method != cmlconstants.PaymentMethodStripe {
			return httpresponse.CreateBadResponse(&c, http.StatusUnauthorized, cmlmessages.PleaseTryAgain, cmlmessages.UnauthorizedForAction)
		}

		// check if there is any negative balance before
		balns, negAmount, err := uc.BalanceService.GetNegBalance(user.ID)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
		}

		if negAmount > float32(body.Amount) {
			// charge more than negative balance
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, fmt.Sprintf("Please charge more than $%f", negAmount))
		}

		stripe.Key = billingParent.Payments.StripeSecretKey
		params := &stripe.ChargeParams{
			Amount:      stripe.Int64(body.Amount * 100),
			Currency:    stripe.String(string(stripe.CurrencyUSD)),
			Description: stripe.String("Add balance"),
		}
		params.SetSource(body.Token)
		ch, err := charge.New(params)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, fmt.Sprintf("Error : %s => Pub Key : %s :: Secret Key : %s", err.Error(), billingParent.Payments.StripePublishableKey, billingParent.Payments.StripeSecretKey))
		}
		//
		logrus.Debugln("Add Balance Stripe : ", *ch)

		bInfo := model.Balance{UserID: user.ID, Amount: float32(body.Amount)}

		if len(balns) > 0 {
			// if there is any ghost balance now is the time to correct it
			blnc := balns[0]
			bInfo.ID = blnc.ID
			bInfo.Consumed = blnc.Consumed
			bInfo.IsGhost = false
			bInfo.IsConsumedFully = false
			bInfo.Expired = false
		}

		// save balance
		uc.BalanceService.Save(
			&bInfo,
		)
		// save payment
		uc.PaymentService.Save(
			&model.Payment{Amount: float32(body.Amount), UserID: user.ID, Comments: "Added by User [Stripe]"},
		)

		// add balance queue
		uc.BillingManager.ProcessAddBalance(user.ID, body.Amount)

		return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Success", "Success")*/
	return nil
}

// this method returns funds availability detail number from dnc
func (uc *UserController) UploadLogo(c echo.Context) error {
	fileHeader, _ := c.FormFile("image")

	userId := c.Param("id")
	// get user object
	user, er1 := uc.UserService.UserObject(userId)
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	logo, err := uc.uploadImage(user, fileHeader)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, err.Error())
	}
	user.Customize.LogoPath = logo
	user.Password = "***********"

	if err := uc.UserService.UpdateLogo(user.ID.Hex(), logo); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, "Logo update failed")
	}

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(user)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", data)
}

// this method returns funds availability detail number from dnc
func (uc *UserController) DeleteLogo(c echo.Context) error {
	userId := c.Param("id")
	user, er1 := uc.UserService.UserObject(userId)
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	user.Customize.LogoPath = ""
	user.Password = "**************"
	if err := uc.UserService.UpdateLogo(user.ID.Hex(), ""); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, "Logo update failed")
	}

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(user)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", data)
}

func (uc *UserController) SendSupportMessage(c echo.Context) error {
	// Stub an user to be populated from the body
	sm := SupportMessage{}

	if err := c.Bind(&sm); err != nil {
		logrus.Infoln("Error: ", err.Error())
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, "Message deliver failure")
	}

	// validate input request body
	if err := c.Validate(sm); err != nil {
		logrus.Infoln(err)
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	go uc.EmailService.SendEmail(&service.EmailOptions{
		FromName:    sm.FirstName + " " + sm.LastName,
		FromEmail:   sm.Email,
		Subject:     "[CML Support Ticket]",
		ToName:      "",
		ToEmail:     "support@callmylist.com",
		TextContent: sm.Message,
		HtmlContent: fmt.Sprintf("<p>%s</p>", sm.Message),
	})

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusCreated, "Success", "Your message has been sent. Someone will contact you shortly.")
}

func (uc *UserController) AddTimeRestrictions(c echo.Context) error {
	userId := c.Param("id")

	// Stub an user to be populated from the body
	tr := model.Restrictions{}

	if err := c.Bind(&tr); err != nil {
		logrus.Infoln("Error: ", err.Error())
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, "Invalid format")
	}

	// validate input request body
	if err := c.Validate(tr); err != nil {
		logrus.Infoln(err)
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	if err := uc.UserService.UpdateTimeRestrictions(userId, tr); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, err.Error())
	}

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(tr)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", data)
}

func (uc *UserController) GenerateApiToken(c echo.Context) error {
	userId := c.Param("id")

	token, err := uc.UserService.GenerateAPIToken(userId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.OperationFailed, err.Error())
	}

	obj := APITokenResponse{Token: token}
	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(obj)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", data)
}

func (uc *UserController) ApiToken(c echo.Context) error {
	userId := c.Param("id")

	// get user object
	user, er1 := uc.UserService.UserObject(userId)
	if er1 != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	obj := APITokenResponse{Token: user.APIToken}

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(obj)
	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", data)
}

func (uc *UserController) sendNewUserEmail(user *model.User, parent *model.User) {
	FromEmail := os.Getenv("SIGNUP_NOTIFICATION_FROM_EMAIL")
	ToEmail := os.Getenv("SIGNUP_NOTIFICATION_TO_EMAIL")

	uc.EmailService.SendEmail(&service.EmailOptions{
		FromName:    user.CompanyName,
		FromEmail:   FromEmail,
		Subject:     "[New User Signup]",
		ToName:      user.FirstName + " " + user.LastName,
		ToEmail:     ToEmail,
		TextContent: fmt.Sprintf("First Name : %s\nLast Name: %s\nEmail : %s\nCompany : %s\nPhone : %s", user.FirstName, user.LastName, user.Email, user.CompanyName, user.Phone),
		HtmlContent: fmt.Sprintf("First Name : %s<br>Last Name: %s<br>Email : %s<br>Company : %s<br>Phone : %s", user.FirstName, user.LastName, user.Email, user.CompanyName, user.Phone),
	})
}

func (uc *UserController) sendVerifyEmail(user *model.User, parent *model.User) error {
	//bParent, err := uc.UserService.UserObject(user.BillingParentID.Hex())
	//if err != nil {
	//	return errors.New("sending email failed")
	//}

	// extract web page url
	url := parent.Account.HomePageUrl
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}

	token := uc.TokenService.GenerateToken(user.Email)
	var endPoint string
	var subject string
	if user.Password == "" {
		endPoint = "auth/verify"
		subject = "Verify Account"
	} else {
		endPoint = "auth/confirm"
		subject = "Confirm Account"
	}

	url = url + endPoint + "?token=" + token

	host, err := host(url)
	if err != nil || host == "" {
		return err
	}
	host = strings.TrimPrefix(host, "www.")

	go func() {
		absPath, _ := filepath.Abs("templates/verify-email-template.gohtml")
		logrus.Infoln(absPath)
		t, err := template.ParseFiles(absPath)
		if err != nil {
			logrus.Errorln("Error : ", err.Error())

			return
		}

		pname := host
		if user.CompanyName != "" {
			pname = parent.CompanyName
		}

		data := struct {
			Name            string
			ProductName     string
			ActionUrl       string
			LogoUrl         string
			EmailAddress    string
			BackgroundColor string
		}{Name: user.FirstName,
			ProductName:     pname,
			ActionUrl:       url,
			LogoUrl:         cmlutils.S3FullPath() + parent.Customize.LogoPath,
			EmailAddress:    user.Email,
			BackgroundColor: parent.Customize.PrimaryColor}

		var tpl bytes.Buffer
		err = t.Execute(&tpl, data)
		if err != nil {
			logrus.Infoln("Error : ", err.Error())

			return
		}
		result := tpl.String()

		uc.EmailService.SendEmail(&service.EmailOptions{
			FromName:    parent.CompanyName,
			FromEmail:   "no-reply@" + host,
			Subject:     subject,
			ToName:      user.FirstName + " " + user.LastName,
			ToEmail:     user.Email,
			TextContent: subject,
			HtmlContent: result,
		})
	}()

	return nil
}

func host(source string) (string, error) {
	u, err := url.Parse(source)
	if err != nil {
		return "", err
	}

	return u.Host, nil
}

func (uc *UserController) uploadImage(user *model.User, file *multipart.FileHeader) (string, error) {

	fileService := uc.FileService
	if !fileService.IsValidImageFileName(file.Filename) {
		return "", errors.New("logo file format is incorrect")
	}

	destPath, err := fileService.SaveFile(file)
	if err != nil {
		return "", errors.New("error saving logo image")
	}
	defer fileService.DeleteFile(destPath)

	// verify size of image
	w, h, err := cmlutils.GetImageDimension(destPath)
	if err != nil {
		return "", errors.New(err.Error())
	}

	if w >= 400 && h >= 400 {
		// image dimensions are correct
		s3Path := "/account/" + user.ID.Hex() + "/" + bson.NewObjectId().Hex() + "/"
		res, err := uc.S3Service.Upload(s3Path, destPath)
		if err != nil {
			return "", errors.New("Error saving logo image")
		}

		return res, nil
	}

	return "", errors.New("Logo dimensions are incorrect")
}

// helper
func AssignJwtToken(c echo.Context, user model.User) error {

	token, err := auth.Token(user)
	if err != nil {
		logrus.Debugln("AssignJwtToken ::: token not generated for user ")
		logrus.Debugln(user)
		return err
	}

	c.Response().Header().Set("x_auth_token", token)

	return nil
}
