package controllerv1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/network/response"
	"gopkg.in/mgo.v2/bson"
)

type BillingController struct {
	Controller
	PaymentService *service.PaymentService
}

// this method creates a campaign
func (bcon *BillingController) BillingHistory(c echo.Context) error {
	cc := c.(model.AppContext)

	// Grab user id for which file is being uploaded
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")

	pageNumber, errPage := strconv.Atoi(page)
	limitNumber, errLimit := strconv.Atoi(limit)

	if errPage != nil || errLimit != nil {
		pageNumber = 1
		limitNumber = 0
	}

	bhList := model.BillingHistories{}
	logrus.Infoln(cc.ID.Hex())
	payment, count, err := bcon.PaymentService.GetPaymentObject(cc.ID.Hex(), pageNumber, limitNumber)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusOK, err.Error(), "")
	}

	logrus.Infoln(payment)

	for _, i := range payment {
		bm := model.BillingHistory{ID: bson.NewObjectId(),
			UserID:     cc.ID,
			Status:     model.BillingHistoryPaid,
			Amount:     i.Amount,
			CreateDate: i.CreateDate,
			UpdateDate: i.UpdateDate,
		}
		bhList = append(bhList, bm)

	}

	pageInfo := model.Page{}
	pageInfo.Page = pageNumber
	pageInfo.Limit = limitNumber
	pageInfo.TotalCount = count

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(bhList)
	pageData, _ := json.Marshal(pageInfo)

	return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusOK, "Billing History List fetched successfully", "Billing History fetched successfully", data, pageData)
}
