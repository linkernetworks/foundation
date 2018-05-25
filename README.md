Mongo
====

[![Build Status](https://travis-ci.org/linkernetworks/gearman.svg?branch=master)](https://travis-ci.org/linkernetworks/gearman)

Mongo is a package integrating gearmand.

# How to use

##### Example

```
cf := *GearmanConfig{}

service := NewFromConfig(cf)

client, err := service.NewClient()

c, err := client.Call(...)
```
