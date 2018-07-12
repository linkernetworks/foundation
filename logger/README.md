logger
===

[![Build Status](https://travis-ci.org/linkernetworks/logger.svg?branch=master)](https://travis-ci.org/linkernetworks/logger)

Logger is a package integrated logger with log rotation.

# How to use

Govendor
```
govendor sync
```

##### Example

```
import "github.com/linkernetworks/foundation/logger"
cf := logger.LoggerConfig
logger.Setup(cf)

logger.Info("my log information")
```
