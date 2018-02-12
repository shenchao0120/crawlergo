package crawlerModel

import (
	"testing"
	"github.com/astaxie/beego/logs"
)

func TestGetDBInstance(t *testing.T) {
	_,err:=GetDBInstance()
	if err!=nil{
		logs.Error(err)
	}

}