package config

import (
	"glmemo/helper/syslog"

	"github.com/go-ini/ini"
)

// GLMEMO ...
var GLMEMO *ini.File

func init() {
	var err error
	GLMEMO, err = ini.Load("./bin/config.ini")
	if err != nil {
		syslog.Clog.Errorln(true, err)
	}

	syslog.Clog.Flag, _ = GLMEMO.Section("log").Key("LogLevel").Int()
}
