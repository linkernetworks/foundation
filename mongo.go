package mongo

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoService struct {
	Url           string
	globalSession *mgo.Session
}

type DataStore struct {
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

func (s *MongoService) NewDataStore() *DataStore {
	return &DataStore{
		Session: s.globalSession.Copy(),
	}
}

func (d *DataStore) Close() {
	d.Session.Close()
}

func (d *DataStore) C(collection string) *mgo.Collection {
	// DB returns a value representing the named database. If name is empty, the database name provided in the dialed URL is used instead.
	return d.Session.DB("").C(collection)
}

func (d *DataStore) FindOne(collection string, key string, value interface{}, r Record) error {
	return d.C(collection).Find(bson.M{key: value}).One(r)
}

func (d *DataStore) Count(collection string, query interface{}) (n int, err error) {
	return d.C(collection).Find(query).Count()
}

func (d *DataStore) FindAll(collection string, query interface{}, records interface{}) error {
	return d.C(collection).Find(query).All(records)
}

func (d *DataStore) FindAllByPage(collection string, query interface{}, sort string, records interface{}, page int, pageSize int) error {
	return d.C(collection).
		Find(query).
		Sort(sort).
		Skip((page - 1) * pageSize).
		Limit(pageSize).All(records)
}

func (d *DataStore) Insert(collection string, r Record) error {
	return d.C(collection).Insert(&r)
}

func (d *DataStore) UpdateBy(collection string, key string, value interface{}, r Record) error {
	return d.C(collection).Update(bson.M{key: value}, r)
}

func (d *DataStore) Upsert(collection string, key string, value interface{}, r Record) (*mgo.ChangeInfo, error) {
	query := bson.M{key: value}
	return d.C(collection).Upsert(query, r)
}

func (d *DataStore) Update(collection string, query bson.M, modifier bson.M) error {
	return d.C(collection).Update(query, modifier)
}

func (d *DataStore) UpdateById(collection string, id bson.ObjectId, update interface{}) error {
	return d.C(collection).UpdateId(id, update)
}

func (d *DataStore) Delete(collection string, key string, value interface{}) error {
	return d.C(collection).Remove(bson.M{key: value})
}

func (d *DataStore) DeleteAll(collection string) (*mgo.ChangeInfo, error) {
	return d.C(collection).RemoveAll(bson.M{})
}

func (d *DataStore) DropCollection(collection string) error {
	return d.C(collection).DropCollection()
}
