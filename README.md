Mongo
====

[![Build Status](https://travis-ci.org/linkernetworks/influxdb.svg?branch=master)](https://travis-ci.org/linkernetworks/influxdb)

Mongo is a package integrating influxdb with influxdb client.

# How to use

##### Example

```

service := *inflxdb.InfluxdbService{
  Url: "db-url",
  Database: "dbname",
}

client := service.NewClient()
defer client.Close()

```
