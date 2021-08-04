package controllerv1

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/gomarkho/sas-rvm-provapi/network/response"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlmessages"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
)

type CampaignHistoryController struct {
	Controller
}

func (chc *CampaignHistoryController) CampaignHistory(c echo.Context) error {

	// Grab user id for which file is being uploaded
	userID := c.Param("id")
	campaignId := c.Param("cid")

	page := c.QueryParam("page")
	limit := c.QueryParam("limit")

	pageNumber, errPage := strconv.Atoi(page)
	limitNumber, errLimit := strconv.Atoi(limit)

	if errPage != nil || errLimit != nil {
		pageNumber = 1
		limitNumber = 0
	}

	url := fmt.Sprintf(cmlutils.VOIPAPIPath()+"user/%s/campaign/%s/campaign-history?page=%d&limit=%d", userID, campaignId, pageNumber, limitNumber)
	resp, err := http.Get(url)
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
