package datasource

import "gopkg.in/mgo.v2"

type DataSource struct {
	Session *mgo.Session
}

// getter for db session
// the reason for creating it is to have a single point of failure
// this approach will make error recovery easy
func (ds *DataSource) DbSession() *mgo.Session {
	return ds.Session
}
