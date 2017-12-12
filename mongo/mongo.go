package mongo

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoService struct {
	Url           string
	globalSession *mgo.Session
}

type Context struct {
	Session *mgo.Session
}

func NewMongoService(url string) *MongoService {
	session, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}
	return &MongoService{
		Url:           url,
		globalSession: session,
	}
}

func (s *MongoService) NewContext() *Context {
	return &Context{
		Session: s.globalSession.Copy(),
	}
}

func (d *Context) Close() {
	d.Session.Close()
}

func (d *Context) C(collection string) *mgo.Collection {
	// DB returns a value representing the named database. If name is empty, the database name provided in the dialed URL is used instead.
	return d.Session.DB("").C(collection)
}

func (d *Context) FindOne(collection string, query bson.M, r Record) error {
	return d.C(collection).Find(query).One(r)
}

func (d *Context) Count(collection string, query interface{}) (n int, err error) {
	return d.C(collection).Find(query).Count()
}

func (d *Context) FindAll(collection string, query interface{}, records interface{}) error {
	return d.C(collection).Find(query).All(records)
}

func (d *Context) FindAllByPage(collection string, query interface{}, sort string, records interface{}, page int, pageSize int) error {
	return d.C(collection).
		Find(query).
		Sort(sort).
		Skip((page - 1) * pageSize).
		Limit(pageSize).All(records)
}

func (d *Context) Insert(collection string, r Record) error {
	return d.C(collection).Insert(&r)
}

func (d *Context) UpdateBy(collection string, key string, value interface{}, r Record) error {
	return d.C(collection).Update(bson.M{key: value}, r)
}

func (d *Context) Upsert(collection string, key string, value interface{}, r Record) (*mgo.ChangeInfo, error) {
	query := bson.M{key: value}
	return d.C(collection).Upsert(query, r)
}

func (d *Context) Update(collection string, query bson.M, modifier bson.M) error {
	return d.C(collection).Update(query, modifier)
}

func (d *Context) UpdateById(collection string, id bson.ObjectId, update interface{}) error {
	return d.C(collection).UpdateId(id, update)
}

func (d *Context) Remove(collection string, key string, value interface{}) error {
	return d.C(collection).Remove(bson.M{key: value})
}

func (d *Context) RemoveAll(collection string) (*mgo.ChangeInfo, error) {
	return d.C(collection).RemoveAll(bson.M{})
}

func (d *Context) DropCollection(collection string) error {
	return d.C(collection).DropCollection()
}
