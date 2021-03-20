package main

import (
	"embed"
	_ "glmemo/helper/database"
	"glmemo/service"
	"runtime"
)

//go:embed web/*
var content embed.FS

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	service.RunService(content)
}
