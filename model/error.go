package model

type (
	// Error represents the structure of our resource
	Error struct {
		Code        int16  `json:"code" bson:"code"`
		Description string `json:"desc" bson:"desc"`
	}
)
