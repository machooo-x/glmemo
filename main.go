package main

import (
	"embed"
	"fmt"
	"glmemo/server"
	"net/http"
	"os"
	"runtime"

	"github.com/labstack/echo"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("marker")
	RunService()
	fmt.Println("marker")

}

//go:embed web/*
var content embed.FS

// RunService 开启服务
func RunService() {

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	// e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
	// 	if username == "admin" && password == "admin" {
	// 		return true, nil
	// 	}
	// 	return false, nil
	// }))
	// html
	webHandler := http.FileServer(http.FS(content))
	e.GET("/*.html", echo.WrapHandler(webHandler))
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/", webHandler)))

	// e.GET("/", func(c echo.Context) error {
	// 	return c.JSON(http.StatusOK, nil)
	// })

	e.GET("/api/testget", server.Testget)
	// e.GET("/api/restore/:mac", service.HTTPPOSTRestore)

	err := e.Start(":1999")
	if err != nil {
		os.Exit(-1)
	}
	// err := e.StartTLS(config.APPConfig.Section("debug").Key("HTTPDebugPort").String(), config.APPConfig.Section("tls").Key("CertFile").String(), config.APPConfig.Section("tls").Key("KeyFile").String())
	// if err != nil {
	// 	os.Exit(-1)
	// }
}
