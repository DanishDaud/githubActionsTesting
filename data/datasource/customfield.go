package datasource

import (
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const ccustomField = "customfield"

type CustomFieldDataSource struct {
	DataSource
}

func (uds *CustomFieldDataSource) Save(obj model.CustomField) (*model.CustomField, error) {

	if obj.ID == "" {
		obj.ID = bson.NewObjectId()
	}

	err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(ccustomField).Insert(obj)

	if err != nil {
		return nil, err
	}

	return &obj, nil

}

func (uds *CustomFieldDataSource) GetById(objectID bson.ObjectId) (*model.CustomField, error) {

	obj := model.CustomField{}
	err := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(ccustomField).Find(bson.M{"_id": objectID}).One(&obj)
	if err != nil {
		return nil, err
	}

	return &obj, nil

}

func (uds *CustomFieldDataSource) List(obj model.CustomFilter) (*[]model.CustomField, int, error) {

	if obj.Page < 1 {
		obj.Page = 1
	}
	if obj.Limit < 1 {
		obj.Limit = 10
	}
	var result []model.CustomField

	if obj.Search == "" {
		query := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(ccustomField).Find(bson.M{"userid": obj.UserId})
		query1 := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(ccustomField).Find(bson.M{"userid": obj.UserId})
		c, err := query.Count()
		if err != nil {
			return nil, 0, err
		}
		query1.Sort("-createDate").Skip((obj.Page - 1) * obj.Limit).Limit(obj.Limit).All(&result)
		if err != nil {
			return nil, 0, err
		}
		return &result, c, nil
	} else {

		query := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(ccustomField).Find(bson.M{"name": bson.RegEx{obj.Search, "i"}})
		query1 := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(ccustomField).Find(bson.M{"name": bson.RegEx{obj.Search, "i"}})
		c, err := query.Count()
		if err != nil {
			return nil, 0, err
		}
		err = query1.Sort("-createDate").Skip((obj.Page - 1) * obj.Limit).Limit(obj.Limit).All(&result)
		if err != nil {
			return nil, 0, err
		}
		return &result, c, nil
	}

	return nil, 0, nil

}
