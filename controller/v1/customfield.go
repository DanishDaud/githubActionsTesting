package controllerv1

import (
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"github.com/gomarkho/sas-rvm-provapi/model"
	httpresponse "github.com/gomarkho/sas-rvm-provapi/network/response"
	"github.com/gomarkho/sas-rvm-provapi/system/cmlmessages"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
	"time"
)

type CustomFieldController struct {
	CustomField *service.CustomFieldService

	Controller
}

func (smCon *CustomFieldController) Save(c echo.Context) error {
	var req model.CustomField
	cc := c.(model.AppContext)
	if err := c.Bind(&req); err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}

	req.UserID = cc.ID
	req.ID = bson.NewObjectId()
	req.CreatedDate = time.Now()
	data, err := smCon.CustomField.Save(req)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}

	datainfo, err := json.Marshal(data)
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}
	return httpresponse.CreateSuccessResponse(&c, http.StatusOK, "success", "success", datainfo)

}

func (smCon *CustomFieldController) List(c echo.Context) error {

	cc := c.(model.AppContext)
	search := c.QueryParam("search")
	limit := c.QueryParam("limit")
	page := c.QueryParam("page")

	pageNumber, errPage := strconv.Atoi(page)
	limitNumber, errLimit := strconv.Atoi(limit)

	if errPage != nil || errLimit != nil {
		pageNumber = 1
		limitNumber = 10
	}

	list, count, err := smCon.CustomField.List(model.CustomFilter{UserId: cc.ID, Search: search, Page: pageNumber, Limit: limitNumber})
	if err != nil {
		return httpresponse.CreateBadResponse(&c, http.StatusBadRequest, cmlmessages.PleaseTryLater, err.Error())
	}

	pageInfo := model.Page{}
	pageInfo.Page = pageNumber
	pageInfo.Limit = limitNumber
	pageInfo.TotalCount = count

	// Marshal provided interface into JSON structure
	data, _ := json.Marshal(list)
	pageData, _ := json.Marshal(pageInfo)
	// Marshal provided interface into JSON structure

	return httpresponse.CreateSuccessResponseWithPageInfo(&c, http.StatusCreated, "Success", "Success", data, pageData)
}
