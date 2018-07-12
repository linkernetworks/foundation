Config
===

[![Build Status](https://travis-ci.org/linkernetworks/config.svg?branch=master)](https://travis-ci.org/linkernetworks/config)

Config is a utility pkg which consumes json file and provide config object.

# How to use

Govendor
```
govendor sync
```

##### Example

```
cf := config.MustRead("example.json")

// use cf
```
