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
	// e.GET("/getfilename", getFilename)
	e.GET("/querytempsave", queryTempSave)
	e.GET("/limitedtimerecord", queryTempSave)
	e.GET("/createtemprecord", createTempRecord)
	e.GET("/gettemprecord", getTempRecord)
	e.GET("/getalltag", getAllTag)
	e.GET("/querybytag", queryByTag)
	e.GET("/querybylike", queryByLike)
	e.GET("/getnetip", getNetIP)
	e.GET("/getToDoList", getToDoList)

	e.POST("/addrecord", addRecord)
	e.POST("/addToDo", addToDo)
	e.POST("/uploadfile", uploadfile)

	// e.PUT("/changerecord", changeRecord)

	e.DELETE("/delrecord", delRecord)
	e.DELETE("/deltempsave", delTempSave)
	e.DELETE("/delToDo", delToDo)

}
