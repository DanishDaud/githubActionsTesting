package datasource

import (
	"github.com/gomarkho/sas-rvm-provapi/model"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"gopkg.in/mgo.v2/bson"
)

const cttsmeta = "ttsmeta"

type TTSMetaDataSource struct {
	DataSource
}

func (uds *TTSMetaDataSource) BulkInsert(obj model.TTSMetaList) error {

	bulk := uds.DbSession().DB(cmlutils.DefaultDatabase()).C(cttsmeta).Bulk()
	var contentArray []interface{}
	for i := range obj {
		contentArray = append(contentArray, obj[i])
	}
	bulk.Insert(contentArray...)
	_, err := bulk.Run()
	if err != nil {
		return err
	}

	return nil

}

func (tnd *TTSMetaDataSource) GetByNumber(number string) (*model.TTSNumber, error) {
	var obj model.TTSNumber
	colQuerier := bson.M{"number": number}
	err := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(cttsmeta).Find(colQuerier).One(&obj)
	if err != nil {
		return nil, err
	}
	return &obj, err

}

func (tnd *TTSMetaDataSource) DeleteAllByID(id bson.ObjectId) error {
	_, err := tnd.DbSession().DB(cmlutils.DefaultDatabase()).C(cttsmeta).RemoveAll(bson.M{"contactlistID": id})
	return err
}
