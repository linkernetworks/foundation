WebSocket
===

[![Build Status](https://travis-ci.org/linkernetworks/websocket.svg?branch=master)](https://travis-ci.org/linkernetworks/websocket)

Websocket is a package integrated websocket by gorilla

# How to use

##### Example

```
service := NewWebSocketService()

if err := service.Run(); err != nil {
  // handler err
}
```
