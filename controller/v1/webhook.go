package controllerv1

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/network/response"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlmessages"
)

type WebhookController struct {
	Controller
	WepayCheckoutService *service.WepayCheckoutService
	BillingManager       *service.BillingManager
	PaymentService       *service.PaymentService
	BalanceService       *service.BalanceService
}

// WePay payment webhook
func (wc *WebhookController) WePayPaymentHeartBeat(c echo.Context) error {
	logrus.Debugln("********* Working *********")
	return nil
}

// WePay payment webhook
func (wc *WebhookController) WePayPaymentProcess(c echo.Context) error {
	logrus.Debugln("********* Working *********")
	bodyBuffer, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}
	logrus.Debugln("Callback => ", string(bodyBuffer))

	items := strings.Split(string(bodyBuffer), "=")
	checkoutId := 0
	if len(items) > 1 && items[0] == "checkout_id" {
		id := items[1]
		checkoutId, err = strconv.Atoi(id)
		if err != nil {
			return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Please provide valid information")
		}
	} else {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, "Please provide valid information")
	}

	cout, err := wc.WepayCheckoutService.WepayCheckoutId(checkoutId)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryAgain, err.Error())
	}

	if !cout.Consumed {
		// save balance
		wc.BalanceService.Save(
			&model.Balance{Amount: cout.Amount},
		)
		// save payment
		wc.PaymentService.Save(
			&model.Payment{Amount: cout.Amount, UserID: cout.UserID, Comments: "Added by User"},
		)

		// add balance to be added in queue
		wc.BillingManager.ProcessAddBalance(cout.UserID, int64(cout.Amount))

		cout.Consumed = true
		wc.WepayCheckoutService.Save(cout)
	}

	return httpresponse.CreateSuccessResponseWithoutData(&c, http.StatusOK, "Success", "Success")
}
