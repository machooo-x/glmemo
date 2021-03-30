package service

import (
	"github.com/labstack/echo"
)

// RegRouter 注册HTTP路由信息
func RegRouter(e *echo.Echo) {

	e.GET("/login", login)
	e.GET("/getrecordlist", getRecordList)
	e.GET("/regist", regist)

	e.GET("/showrecord", showRecord)
	e.GET("/getfilename", getFilename)
	e.GET("/querytempsave", queryTempSave)

	e.POST("/addrecord", addRecord)
	e.POST("/uploadfile", uploadfile)

	// e.PUT("/changerecord", changeRecord)

	e.DELETE("/delrecord", delRecord)
	e.DELETE("/deltempsave", delTempSave)

}
