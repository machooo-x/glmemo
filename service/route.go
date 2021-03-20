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

	e.POST("/addrecord", addRecord)

	e.PUT("/changerecord", changeRecord)

	e.DELETE("/delrecord", delRecord)

}
