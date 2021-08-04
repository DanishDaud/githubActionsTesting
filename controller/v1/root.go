package controllerv1

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/network/response"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlmessages"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
)

type RootController struct {
	TokenService *service.Token
	EmailService *service.Email
	UserService  *service.UserService
	Controller
}

type SendEmailOptions struct {
	User     *model.User
	Parent   *model.User
	File     string
	EndPoint string
	Subject  string
}

// data structure to get credentials from json object
type (
	nameObject struct {
		Name string `json:"name"`
	}
)

// CreateUser creates a new user resource
func (rootCon *RootController) GetAssetHomePath(c echo.Context) error {

	obj := nameObject{}
	obj.Name = cmlutils.S3ProxyPath()

	data, _ := json.Marshal(obj)

	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "Host Name returned successfully", "", data)
}

func (rootCon *RootController) ForgotPasswordVerify(c echo.Context) error {
	t := c.QueryParam("token")

	newPassword := c.QueryParam("newPassword")

	if t == "" || newPassword == "" || len(newPassword) < 8 {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Invalid input")
	}

	email, err := rootCon.TokenService.Verify(t)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Invalid token")
	}

	uService := rootCon.UserService

	user, err := uService.UserObjectWithEmailAndBillingParent(email, true)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "you are not authorized to perform this action")
	}

	if err := uService.ChangePassword(user.ID.Hex(), newPassword); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	user.Password = "**************************"

	// if account is locked
	if user.IsLocked {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserAccountLocked)
	}

	// Marshal provided interface into JSON structure
	uj, _ := json.Marshal(*user)

	// generate jwt token
	if err := AssignJwtToken(c, *user); err != nil {
		logrus.Infoln("token not generated")
	}

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", uj)
}

func (rootCon *RootController) VerifyEmailWithPassword(c echo.Context) error {
	t := c.QueryParam("token")
	password := c.QueryParam("password")

	if t == "" || len(password) < 8 {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Invalid values")
	}

	email, err := rootCon.TokenService.Verify(t)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Invalid token")
	}

	uService := rootCon.UserService

	user, err := uService.UserObjectWithEmailAndBillingParent(email, false)
	if err != nil {
		user, err = uService.UserObjectWithEmailAndBillingParent(email, true)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
		}
	}

	user.IsVerified = true

	if err := uService.VerifyUser(user.ID.Hex()); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	if err := uService.ChangePassword(user.ID.Hex(), password); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	user.Password = "**************************"

	// if account is locked
	if user.IsLocked {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserAccountLocked)
	}

	// Marshal provided interface into JSON structure
	uj, _ := json.Marshal(*user)

	// generate jwt token
	if err := AssignJwtToken(c, *user); err != nil {
		logrus.Infoln("token not generated")
	}

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", uj)
}

func (rootCon *RootController) ConfirmEmail(c echo.Context) error {
	t := c.QueryParam("token")

	if t == "" {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Invalid token")
	}

	email, err := rootCon.TokenService.Verify(t)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Invalid token")
	}

	uService := rootCon.UserService

	user, err := uService.UserObjectWithEmailAndBillingParent(email, false)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	user.IsVerified = true

	if err := uService.VerifyUser(user.ID.Hex()); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	user.Password = "**************************"

	// if account is locked
	if user.IsLocked {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserAccountLocked)
	}

	// Marshal provided interface into JSON structure
	uj, _ := json.Marshal(*user)

	// generate jwt token
	if err := AssignJwtToken(c, *user); err != nil {
		logrus.Infoln("token not generated")
	}

	return httpresponse.CreateSuccessResponse(&c, http.StatusCreated, "Success", "Success", uj)
}

func (rootCon *RootController) ResendVerify(c echo.Context) error {
	// password is required
	userId := c.QueryParam("userId")
	uService := rootCon.UserService

	user, err := uService.UserObject(userId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.UserDoesNotExist)
	}

	if user.IsVerified {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "user is already verified")
	}

	var filePath string
	if os.Getenv("ENV") == "local" {

		filePath = "templates/forgot-password-template.gohtml"
	} else {
		filePath = "../templates/verify-email-template.gohtml"
	}

	var endPoint string
	var subject string
	if user.Password == "" {
		endPoint = "verify"
		subject = "Verify Account"
	} else {
		endPoint = "confirm"
		subject = "Confirm Account"
	}

	if err := rootCon.sendEmail(&SendEmailOptions{User: user,
		Parent:   user,
		EndPoint: endPoint,
		File:     filePath,
		Subject:  subject,
	}); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusCreated, "Success", "In few minutes you will receive email with link to verify your account")
}

func (rootCon *RootController) SendForgotPasswordEmail(c echo.Context) error {

	email := c.QueryParam("email")
	//billingParent := c.QueryParam("ParentId")
	if email == "" {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "email is compulsory")
	}

	uService := rootCon.UserService

	user, err := uService.UserObjectWithEmailAndBillingParent(email, true)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "you are not authorized to perform this action")
	}

	// check if user is verified
	//if !user.IsVerified || !user.Role.IsAuthorized() || user.GMailAccountId != "" {
	//	return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "you are not authorized to perform this action")
	//}

	//	bParent, err := uService.UserObject(user.BillingParentID.Hex())
	//	if err != nil {
	//		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	//	}

	var filePath string
	if os.Getenv("ENV") == "local" {
		filePath = "templates/forgot-password-template.gohtml"
	} else {
		filePath = "templates/forgot-password-template.gohtml"
	}

	if err := rootCon.sendEmail(&SendEmailOptions{User: user,
		Parent:   user,
		EndPoint: "auth/reset",
		File:     filePath,
		Subject:  "Forgot Password",
	}); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, cmlmessages.OperationFailed)
	}

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusCreated, "Success", "In few minutes you will receive email with link to use to recover your password")
}

func (rootCon *RootController) sendEmail(options *SendEmailOptions) error {
	// extract web page url
	url := options.Parent.Account.HomePageUrl
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}

	token := rootCon.TokenService.GenerateToken(options.User.Email)
	logrus.Infoln(token)
	url = url + options.EndPoint + "?token=" + token

	host, err := host(url)
	if err != nil || host == "" {
		return err
	}
	host = strings.TrimPrefix(host, "www.")

	pname := host
	if options.Parent.CompanyName != "" {
		pname = options.Parent.CompanyName
	}

	go func() {
		absPath, _ := filepath.Abs(options.File)
		logrus.Errorln(absPath)
		t, err := template.ParseFiles(absPath)
		if err != nil {
			logrus.Errorln("Error : ", err.Error())
			return
		}

		data := struct {
			Name            string
			ProductName     string
			ActionUrl       string
			LogoUrl         string
			EmailAddress    string
			BackgroundColor string
		}{Name: options.User.FirstName,
			ProductName:     pname,
			ActionUrl:       url,
			LogoUrl:         cmlutils.S3FullPath() + options.Parent.Customize.LogoPath,
			EmailAddress:    options.User.Email,
			BackgroundColor: options.Parent.Customize.PrimaryColor}

		var tpl bytes.Buffer
		err = t.Execute(&tpl, data)
		if err != nil {
			logrus.Infoln("Error : ", err.Error())
			return
		}
		result := tpl.String()

		rootCon.EmailService.SendEmail(&service.EmailOptions{
			FromName:    options.Parent.CompanyName,
			FromEmail:   "no-reply@" + host,
			Subject:     options.Subject,
			ToName:      options.User.FirstName + " " + options.User.LastName,
			ToEmail:     options.User.Email,
			TextContent: options.Subject,
			HtmlContent: result,
		})
	}()

	return nil
}
