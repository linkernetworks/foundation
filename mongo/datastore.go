package mongo

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type DataStore struct {
	Context    *Session
	Session    *mgo.Session
	Collection string
	C          *mgo.Collection
}

func NewDataStore(context *Session, collection string) *DataStore {
	c := context.Session.DB("").C(collection)
	return &DataStore{context, context.Session, collection, c}
}

func (d *DataStore) FindOne(q bson.M, r Record) error {
	return d.C.Find(q).One(r)
}

func (d *DataStore) Count(q interface{}) (n int, err error) {
	return d.C.Find(q).Count()
}

func (d *DataStore) FindAll(q interface{}, records interface{}) error {
	return d.C.Find(q).All(records)
}

func (d *DataStore) FindAllByPage(q interface{}, sort string, records interface{}, page int, pageSize int) error {
	return d.C.
		Find(q).
		Sort(sort).
		Skip((page - 1) * pageSize).
		Limit(pageSize).All(records)
}

func (d *DataStore) Insert(r Record) error {
	return d.C.Insert(&r)
}

func (d *DataStore) UpdateBy(key string, value interface{}, r Record) error {
	return d.C.Update(bson.M{key: value}, r)
}

func (d *DataStore) Upsert(key string, value interface{}, r Record) (*mgo.ChangeInfo, error) {
	q := bson.M{key: value}
	return d.C.Upsert(q, r)
}

func (d *DataStore) Update(q bson.M, modifier bson.M) error {
	return d.C.Update(q, modifier)
}

func (d *DataStore) UpdateById(id bson.ObjectId, update interface{}) error {
	return d.C.UpdateId(id, update)
}

func (d *DataStore) Remove(key string, value interface{}) error {
	return d.C.Remove(bson.M{key: value})
}

func (d *DataStore) RemoveAll() (*mgo.ChangeInfo, error) {
	return d.C.RemoveAll(bson.M{})
}

func (d *DataStore) DropCollection() error {
	return d.C.DropCollection()
}
