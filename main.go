package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	_ "mbook/routers"
	_ "mbook/sysinit"
)

func main() {
	logs.SetLogger(logs.AdapterFile,`{"filename":"project.log","level":5,"maxlines":0,"maxsize":0,"daily":true,"maxdays":10,"color":true}`)
	beego.Run()
}
