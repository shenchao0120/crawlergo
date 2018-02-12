package crawlerModel

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"sync"
	"github.com/astaxie/beego/logs"
	"chaoshen.com/crawlergo/crawler/config"
	"errors"
)

var db *sql.DB

var once sync.Once

func GetDBInstance() (*sql.DB,error) {
	var err error = nil
	once.Do(func() {
		db,err=sql.Open(config.GetConfig().Database.DriverName,config.GetDBConnectString())
		if err!=nil{
			logs.Error("Open connect to mysql error.")
		}
		err=db.Ping()
		if err!=nil{
			logs.Error("Ping mysql database error.")
		}
	})
	return db,err
}

type RequestInfo struct {
	id uint64
	url string
	domain string
	legal bool
	create string
}


func InsertRequestInfo(req *RequestInfo) (sql.Result,error){
	if req == nil || req.url=="" {
		return nil,errors.New("The input url cannot be nil.")
	}
	return nil,nil
}




