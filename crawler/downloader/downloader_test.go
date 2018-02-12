package downloader

import (
	"testing"
	"net/http"
	"fmt"
	"chaoshen.com/crawlergo/crawler/basic"
	"os"
	"bytes"
	"net/url"
	"github.com/astaxie/beego/logs"
)

func TestPageDownloader(t *testing.T) {
	downloader:=NewPageDownloader(nil)
	req,err:=http.NewRequest("GET","http://www.cbhb.com.cn",nil)
	if err!=nil {
		fmt.Println("New Request error",err)
	}
	dlReq:=basic.NewDownloadRequest(1,req,1)
	dlResp,err:=downloader.Download(dlReq)
	if err!=nil {
		fmt.Println("New Request error",err)
	}
	file,err:=os.Create("./output1.txt")
	defer file.Close()
	if err!=nil {
		fmt.Println("Create file error",err)
	}

	dlResp.HttpResp().Write(file)
	fmt.Printf("The request id %d",dlResp.GetID())

}


func TestNewPageDownloaderPool(t *testing.T) {
	dlPool,err:=NewPageDownloaderPool(5)
	if err!=nil {
		t.Error("new pool err:",err)
	}
	dl,err:=dlPool.Take()
	fmt.Println("pool used",dlPool.Used())
	dl2,err:=dlPool.Take()
	fmt.Println("pool used",dlPool.Used())
	dl3,err:=dlPool.Take()
	fmt.Println("pool used",dlPool.Used())
	dlPool.Return(dl2)
	dlPool.Return(dl3)
	req,err:=http.NewRequest("GET","http://www.baidu.com.cn",nil)
	if err!=nil {
		fmt.Println("New Request error",err)
	}
	dlReq:=basic.NewDownloadRequest(1,req,1)
	dlResp,err:=dl.Download(dlReq)
	if err!=nil {
		fmt.Println("New Response error",err)
	}
	bf:=new(bytes.Buffer)
	dlResp.HttpResp().Write(bf)
	fmt.Println("The response is ",dlResp.GetID())
	dlPool.Return(dl)
	fmt.Println("pool used",dlPool.Used())
}


func TestBlogURlTest (t *testing.T){
	urlString:="http://blog.csdn.net/T7SFOKzorD1JAYMSFk4"
	url,_:=url.Parse(urlString)
	logs.Info("schema:%s",url.Host)

}