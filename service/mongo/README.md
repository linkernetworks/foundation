Mongo
====

[![Build Status](https://travis-ci.org/linkernetworks/mongo.svg?branch=master)](https://travis-ci.org/linkernetworks/mongo)

Mongo is a package integrating mongoDB by mgo.

# How to use

##### Example

```
const mongoUrl = "mongodb://you.mongo.db.ip:port/db-name"
mongoService := mongo.New(mongoUrl)

session := mongoService.NewSession()
defer session.Close()

entity := new(Entity)
query := bson.M{}
if err := session.C(collectionName).Find(query).All(&entity); if err != nil {
  // Handler error
}
fmt.Println(entity)
```
