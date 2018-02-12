package main

import (
	"bytes"
	"chaoshen.com/crawlergo/crawler/basic"
	"chaoshen.com/crawlergo/crawler/pageParser"
	"chaoshen.com/crawlergo/crawler/pipeline"
	sched  "chaoshen.com/crawlergo/crawler/scheduler"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/astaxie/beego/logs"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"errors"
	"github.com/astaxie/beego"
)

const maxCrawler = 2

func main() {
	basic.LoggerInit()
	beego.SetLogger("file", `{"filename":"crawler/demo/logs/crawler.log"}`)
	logs.SetLevel(logs.LevelInformational)

	channelConfig := basic.NewChannelConfig(20, 20, 10, 5)
	poolBaseConfig := basic.NewPoolBaseConfig(20, 20)

	httpClientGen := func() *http.Client {
		return &http.Client{
			Timeout:2*time.Second,
		}
	}
	respParsers := getResponseParsers()
	processor := getItemProcessor()

	initUrl := "http://www.csdn.net"
	req, err := http.NewRequest("GET", initUrl, nil)
	if err != nil {
		logs.Error("Init first Request error:%s\n", err)
		return
	}

	scheduler, err := sched.NewScheduler(maxCrawler,
		channelConfig,
		poolBaseConfig,
		httpClientGen,
		respParsers,
		processor,
	)

	if err != nil {
		logs.Error("Fatal Error %s", err)
		return
	}

	intervalNs := 10 * time.Millisecond

	scheduler.Start(req)

	flagChan := sched.Monitoring(scheduler, intervalNs, false)
	if err != nil {
		logs.Error("Start Monitor error:%s", err)
		return
	}
	<-flagChan
}
func getResponseParsers() []pageParser.ParseResponse {
	parsers := []pageParser.ParseResponse{
		parseForATag,
	}
	return parsers
}

func getItemProcessor() []itemproc.ProcessItem {
	processor := []itemproc.ProcessItem{
		processItemPrint,
	}
	return processor
}

func parseForATag(httpResp *http.Response, respDepth uint32) ([]basic.BaseData, []error) {
	if httpResp.StatusCode != 200 {
		err := errors.New(
			fmt.Sprintf("Unsupported status code %d. (httpResponse=%v)", httpResp))
		return nil, []error{err}
	}
	defer func() {
		if httpResp.Body != nil {
			httpResp.Body.Close()
		}
	}()

	bodyStr := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		_, err := httpResp.Body.Read(buf) // 为这次读出的数据大小
		if err != nil && err != io.EOF {
			return nil, []error{err}
		}
		if err == io.EOF { /* && n != 0*/
			bodyStr = append(bodyStr, buf...)
			break
		}
		bodyStr = append(bodyStr, buf...)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(bodyStr))
	if err != nil {
		return nil, []error{err}
	}

	dataList := make([]basic.BaseData, 0)
	errorList := make([]error, 0)
	doc.Find("a").Each(func(index int, sel *goquery.Selection) {
		href, isExist := sel.Attr("href")
		if !isExist || href == "" || href == "#" || href == "/" {
			return
		}
		href = strings.TrimSpace(href)
		lowerHref := strings.ToLower(href)
		if href == "" && strings.HasPrefix(lowerHref, "javascript") {
			return
		}
		url, err := url.Parse(href)
		if err != nil {
			errorList = append(errorList, err)
			return
		}
		if !url.IsAbs() {
			url = httpResp.Request.URL.ResolveReference(url)
		}
		newReq, err := http.NewRequest("GET", url.String(), nil)
		if err != nil {
			errorList = append(errorList, err)
			return
		}
		dataList = append(dataList, basic.NewDownloadRequest(0, newReq, respDepth))
		logs.Debug("Find new http request:%s.", url.String())
	})

	imap := make(map[string]interface{})
	doc.Find("title").Each(func(index int, sel *goquery.Selection) {
		text := strings.TrimSpace(sel.Text())
		if text != "" {
			if _, ok := imap["title"]; ok {
				return
			}
			imap["title"] = text
			imap["url"] = httpResp.Request.URL.String()
			imap["body"] = string(bodyStr)
			item := basic.ItemMap(imap)
			dataList = append(dataList, item)
		}
	})

	return dataList, errorList
}

func processItemPrint(itemMap basic.ItemMap) (basic.ItemMap, error) {
	//logs.Info("ItemMap:")
	var title string
	var url string
	var body string
	for k, v := range itemMap {
		switch k {
		case "title":
			title, _ = v.(string)
		case "url":
			url, _ = v.(string)
		case "body":
			body, _ = v.(string)

		}
	}
	logs.Info("Find Item title:%s,url:%s", title, url)
	if strings.Contains(url, "blog") {
		logs.Critical("Find blog title:%s,url:%s,body:%d", title, url, len(body))
	}
	return itemMap, nil

}
