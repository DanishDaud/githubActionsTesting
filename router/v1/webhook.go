package apirouterv1

import (
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2"
)

func RouteWebhookApis(acc *echo.Group, res *echo.Group, session *mgo.Session) {
	// initialize service of user
	//checkoutService := service.WepayCheckoutService{Service: service.Service{Session: session}}

	// initialize service of balance
	//	balanceService := service.BalanceService{Service: service.Service{Session: session}}

	// initialize service of payment
	//	paymentService := service.PaymentService{Service: service.Service{Session: session}}

	// billing manager
	//bm := service.NewBillingManager()

	// initialize user controller
	//wc := controllerv1.WebhookController{
	//	WepayCheckoutService: &checkoutService,
	//	BillingManager: bm,
	//	BalanceService: &balanceService,
	//	PaymentService: &paymentService,
	//}

	//// Create a new user
	//acc.POST("/wepay/payments", wc.WePayPaymentProcess)
	//
	//// Create a new user
	//acc.GET("/wepay/payments", wc.WePayPaymentHeartBeat)
}
