package model

import "gopkg.in/mgo.v2/bson"

type (
	Page struct {
		Page       int `json:"page"`
		Limit      int `json:"limit"`
		TotalCount int `json:"totalCount"`
	}
)
type DncNumber struct {
	FileName      string        `json:"filename"`
	ContactListId bson.ObjectId `json:"contactlistid"`
	GroupId       bson.ObjectId `json:"groupid"`
	Number        string        `json:"number"`
}
