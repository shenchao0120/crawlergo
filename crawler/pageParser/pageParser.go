package pageParser

import (
	"chaoshen.com/crawlergo/crawler/basic"
	"chaoshen.com/crawlergo/crawler/util"
	"net/http"
	"errors"
	"github.com/astaxie/beego/logs"
	"fmt"
)

type ParseResponse func(httpResp *http.Response, respDepth uint32) ([]basic.BaseData, []error)

var idGenerator = util.NewIdGenerator()

type PageParser interface {
	Id() uint32
	ParsePage(respParsers []ParseResponse, respond *basic.DownloadRespond) ([]basic.BaseData, []error)
}

type PageParserImpl struct {
	id uint32
}

func NewPageParser() PageParser {
	return &PageParserImpl{id: idGenerator.GetUint32Id()}
}

func (ppi *PageParserImpl) Id() uint32 {
	return ppi.id
}
func (ppi *PageParserImpl) ParsePage(
	respParsers []ParseResponse,
	respond *basic.DownloadRespond) ([]basic.BaseData, []error) {
	if respond == nil {
		return nil ,[]error{errors.New("The DownloadRespond is nil.")}
	}
	if respParsers ==nil {
		return nil ,[]error{errors.New("The respParsers is nil.")}
	}
	httpResp:= respond.HttpResp()
	if httpResp==nil {
		return nil ,[]error{errors.New("The httpResp is nil.")}
	}
	reqUrl:=httpResp.Request.URL
	logs.Info("Begin parse the response (reqUrl=%s)... \n", reqUrl)

	reqDepth:=respond.Depth()

	dataList:=make([]basic.BaseData,0)
	errorList:=make([]error,0)


	for i,respParser:=range respParsers{
		if respParser==nil {
			err:=errors.New(fmt.Sprintf("The document parser [%d] is invalid!",i))
			errorList=append(errorList,err)
		}
		pDataList,pErrorList:=respParser(httpResp,reqDepth)
		for _,data:=range pDataList {
			dataList=appendDataList(dataList,data,reqDepth)
		}

		for _,err:= range pErrorList{
			errorList=appendErrorList(errorList,err)
		}
	}
	return dataList,errorList
}

func appendDataList(dataList []basic.BaseData,data basic.BaseData,depth uint32) []basic.BaseData{
	if data==nil {
		return dataList
	}
	req,ok:=data.(*basic.DownloadRequest)
	if !ok{
		return append(dataList,data)
	}
	if req.Depth()!=depth+1 {
		req=basic.NewDownloadRequest(req.GetID(),req.HttpReq(),depth+1)
	}
	return append(dataList,req)
}

func appendErrorList(errorList []error ,err error) []error{
		if err == nil {
			return errorList
		}
		return append(errorList,err)
}
