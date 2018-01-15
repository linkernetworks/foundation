package mongo

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Session struct {
	Session *mgo.Session
}

func (c *Session) NewDataStore(collection string) *DataStore {
	return NewDataStore(c, collection)
}

func (c *Session) Close() {
	c.Session.Close()
}

func (c *Session) C(collection string) *mgo.Collection {
	// DB returns a value representing the named database. If name is empty, the database name provided in the dialed URL is used instead.
	return c.Session.DB("").C(collection)
}

func (c *Session) FindOne(collection string, query bson.M, r Record) error {
	return c.C(collection).Find(query).One(r)
}

func (c *Session) Count(collection string, query interface{}) (n int, err error) {
	return c.C(collection).Find(query).Count()
}

func (c *Session) FindAll(collection string, query interface{}, records interface{}) error {
	return c.C(collection).Find(query).All(records)
}

func (c *Session) FindAllByPage(collection string, query interface{}, sort string, records interface{}, page int, pageSize int) error {
	return c.C(collection).
		Find(query).
		Sort(sort).
		Skip((page - 1) * pageSize).
		Limit(pageSize).All(records)
}

func (c *Session) Insert(collection string, r Record) error {
	return c.C(collection).Insert(&r)
}

func (c *Session) UpdateBy(collection string, key string, value interface{}, r Record) error {
	return c.C(collection).Update(bson.M{key: value}, r)
}

func (c *Session) Upsert(collection string, key string, value interface{}, r Record) (*mgo.ChangeInfo, error) {
	query := bson.M{key: value}
	return c.C(collection).Upsert(query, r)
}

func (c *Session) Update(collection string, query bson.M, modifier bson.M) error {
	return c.C(collection).Update(query, modifier)
}

func (c *Session) UpdateAll(collection string, sel bson.M, update bson.M) (*mgo.ChangeInfo, error) {
	return c.C(collection).UpdateAll(sel, update)
}

func (c *Session) UpdateById(collection string, id bson.ObjectId, update interface{}) error {
	return c.C(collection).UpdateId(id, update)
}

func (c *Session) Remove(collection string, key string, value interface{}) error {
	return c.C(collection).Remove(bson.M{key: value})
}

func (c *Session) RemoveAll(collection string) (*mgo.ChangeInfo, error) {
	return c.C(collection).RemoveAll(bson.M{})
}

func (c *Session) DropCollection(collection string) error {
	return c.C(collection).DropCollection()
}
