package server

import (
	"net/http"

	"github.com/labstack/echo"
)

// Testget ...
func Testget(c echo.Context) error {
	str := "Hello World!"
	strSlice := []string{}
	for i := 0; i < 20; i++ {
		strSlice = append(strSlice, str)
	}
	return c.JSON(http.StatusOK, strSlice)
}
