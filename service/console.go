package service

import (
	"embed"
	"glmemo/helper/syslog"
	"net/http"
	"os"

	"github.com/labstack/echo"
)

// RunService 开启服务
func RunService(content embed.FS) {
	syslog.Clog.Infoln(true, "good life memo 初始化中 ...")

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	webHandler := http.FileServer(http.FS(content))
	e.GET("/*.html", echo.WrapHandler(webHandler))
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/", webHandler)))
	/* file */
	e.Static("/data", "data")
	RegRouter(e)
	err := e.Start(":80")
	if err != nil {
		os.Exit(-1)
	}
}
