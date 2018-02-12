package basic

import "net/http"

type BaseData interface {
	IsValid() bool // Is this data valid.
}

type DownloadRequest struct {
	id uint64
	httpRequest *http.Request
	depth uint32
}

func NewDownloadRequest(id uint64,httpRequest *http.Request,depth uint32) *DownloadRequest{
	return &DownloadRequest{id:id,httpRequest:httpRequest,depth:depth}
}

func (req *DownloadRequest)HttpReq() *http.Request{
	return req.httpRequest
}

func (req *DownloadRequest)IsValid() bool{
	return req.httpRequest!=nil && req.depth!=0
}


func (req *DownloadRequest)Depth() uint32{
	return req.depth
}

func (req *DownloadRequest)GetID() uint64{
	return req.id
}

type DownloadRespond struct {
	id uint64
	httpResponse *http.Response
	depth uint32
}

func NewDownloadResponse(id uint64, httpResponse *http.Response,depth uint32) *DownloadRespond{
	return &DownloadRespond{id:id,httpResponse:httpResponse,depth:depth}
}


func (resp *DownloadRespond)HttpResp() *http.Response{
	return resp.httpResponse
}

func (resp *DownloadRespond)Depth() uint32{
	return resp.depth
}

func (resp *DownloadRespond)GetID() uint64{
	return resp.id
}



// 条目。
type ItemMap map[string]interface{}

// 数据是否有效。
func (item ItemMap) IsValid() bool {
	return item != nil
}

