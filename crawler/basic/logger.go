package basic

import "github.com/astaxie/beego/logs"

func LoggerInit(){
	logs.SetLogFuncCall(true)
	logs.SetLogger("console")
}