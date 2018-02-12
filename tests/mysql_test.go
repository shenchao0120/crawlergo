package test

import (
	"testing"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/astaxie/beego/logs"
)

func TestMysqlConnect(t *testing.T){
	db,err:=sql.Open("mysql","crawler:crawler@tcp(127.0.0.1:3306)/crawlergo?charset=utf8")
	CheckErr(err)
	_, err = db.Exec("INSERT INTO userinfo (username, departname, created) VALUES (?, ?, ?)","lily","销售","2016-06-21")
	CheckErr(err)
	var username, departname, created string
	err = db.QueryRow("SELECT username,departname,created FROM userinfo WHERE uid=?", 7).Scan(&username, &departname, &created)
	CheckErr(err)
	logs.Info("Result:%s,%s,%s",username,departname,created)
}

func CheckErr(err error){
	if err != nil {
		logs.Error("Error:%s ",err)
	}
	return
}