package mongo

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Context struct {
	Session *mgo.Session
}

func (c *Context) NewDataStore(collection string) *DataStore {
	return NewDataStore(c, collection)
}

func (c *Context) Close() {
	c.Session.Close()
}

func (c *Context) C(collection string) *mgo.Collection {
	// DB returns a value representing the named database. If name is empty, the database name provided in the dialed URL is used instead.
	return c.Session.DB("").C(collection)
}

func (c *Context) FindOne(collection string, query bson.M, r Record) error {
	return c.C(collection).Find(query).One(r)
}

func (c *Context) Count(collection string, query interface{}) (n int, err error) {
	return c.C(collection).Find(query).Count()
}

func (c *Context) FindAll(collection string, query interface{}, records interface{}) error {
	return c.C(collection).Find(query).All(records)
}

func (c *Context) FindAllByPage(collection string, query interface{}, sort string, records interface{}, page int, pageSize int) error {
	return c.C(collection).
		Find(query).
		Sort(sort).
		Skip((page - 1) * pageSize).
		Limit(pageSize).All(records)
}

func (c *Context) Insert(collection string, r Record) error {
	return c.C(collection).Insert(&r)
}

func (c *Context) UpdateBy(collection string, key string, value interface{}, r Record) error {
	return c.C(collection).Update(bson.M{key: value}, r)
}

func (c *Context) Upsert(collection string, key string, value interface{}, r Record) (*mgo.ChangeInfo, error) {
	query := bson.M{key: value}
	return c.C(collection).Upsert(query, r)
}

func (c *Context) Update(collection string, query bson.M, modifier bson.M) error {
	return c.C(collection).Update(query, modifier)
}

func (c *Context) UpdateById(collection string, id bson.ObjectId, update interface{}) error {
	return c.C(collection).UpdateId(id, update)
}

func (c *Context) Remove(collection string, key string, value interface{}) error {
	return c.C(collection).Remove(bson.M{key: value})
}

func (c *Context) RemoveAll(collection string) (*mgo.ChangeInfo, error) {
	return c.C(collection).RemoveAll(bson.M{})
}

func (c *Context) DropCollection(collection string) error {
	return c.C(collection).DropCollection()
}
