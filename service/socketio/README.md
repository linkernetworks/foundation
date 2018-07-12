logger
===

[![Build Status](https://travis-ci.org/linkernetworks/socketio.svg?branch=master)](https://travis-ci.org/linkernetworks/socketio)

Socketio is a package integrated socketio with go-socket.io

# How to use

##### Example

```
cf := config.SocketioConfig{...}

service := socketio.New(cf)

client, ok := service.GetClient()
defer client.Stop()

```
