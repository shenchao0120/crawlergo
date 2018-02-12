package downloader

import (
	"net/http"
	"chaoshen.com/crawlergo/crawler/basic"
	"chaoshen.com/crawlergo/crawler/util"
	"github.com/astaxie/beego/logs"
)

// 日志记录器。

// ID生成器。
var idGenerator util.IdGenerator = util.NewIdGenerator()

// 生成并返回ID。
func genDownloaderId() uint32 {
	return idGenerator.GetUint32Id()
}

// 网页下载器的接口类型。
type PageDownloader interface {
	Id() uint32                                        // 获得ID。
	Download(req *basic.DownloadRequest) (*basic.DownloadRespond, error) // 根据请求下载网页并返回响应。
}

// 创建网页下载器。
func NewPageDownloader(client *http.Client) PageDownloader {
	id := genDownloaderId()
	if client == nil {
		client = &http.Client{}
	}
	return &pageDownloaderImpl{
		id:         id,
		httpClient: *client,
	}
}

// 网页下载器的实现类型。
type pageDownloaderImpl struct {
	id         uint32      // ID。
	httpClient http.Client // HTTP客户端。
}

func (dl *pageDownloaderImpl) Id() uint32 {
	return dl.id
}

func (dl *pageDownloaderImpl) Download(req *basic.DownloadRequest) (*basic.DownloadRespond, error) {
	httpReq := req.HttpReq()
	logs.Info("Do the request (url=%s)... \n", httpReq.URL)
	httpResp, err := dl.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	return basic.NewDownloadResponse(req.GetID(),httpResp, req.Depth()), nil
}
