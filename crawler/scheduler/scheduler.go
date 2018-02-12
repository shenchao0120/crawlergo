package scheduler

import (
	"chaoshen.com/crawlergo/crawler/basic"
	"chaoshen.com/crawlergo/crawler/downloader"
	"chaoshen.com/crawlergo/crawler/pageParser"
	"chaoshen.com/crawlergo/crawler/pipeline"
	"chaoshen.com/crawlergo/crawler/util"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	SCHEDULER_STATUS_ALLOCATE = iota
	SCHEDULER_STATUS_READY
	SCHEDULER_STATUS_STARTING
	SCHEDULER_STATUS_RUNNING
	SCHEDULER_STATUS_CLOSED
	SCHEDULER_STATUS_FATAL_ERROR
)

const (
	DOWNLOADER_CODE   = "downloader"
	PARSER_CODE       = "parser"
	ITEMPIPELINE_CODE = "item_pipeline"
	SCHEDULER_CODE    = "scheduler"
)

type Scheduler interface {
	Start(initRequest *http.Request) error
	Stop() error
	Status() uint
	ErrorChan() <-chan error
	Idle() bool
	AddPermitDomain(host string) error
	Summary() SchedSummary
}

type schedulerImpl struct {
	sync.RWMutex
	mapLock 	   sync.RWMutex
	channelConfig  basic.ChannelConfig
	poolBaseConfig basic.PoolBaseConfig
	crawMaxDepth   uint32
	urlMap         map[string]bool
	status         uint32
	acceptDomain   map[string]struct{}
	channelManager util.ChannelManager
	stopSign       util.StopSign
	dlPool         downloader.PageDownloaderPool
	parserPool     pageParser.PageParserPool
	pageParsers    []pageParser.ParseResponse
	itemPipeline   itemproc.ItemPipeline
	reqCache       basic.RequestCache
}

func NewScheduler(rawMaxDepth uint32,
	channelConfig basic.ChannelConfig,
	poolBaseConfig basic.PoolBaseConfig,
	httpClientGenerator downloader.GenHttpClient,
	pageParsers []pageParser.ParseResponse,
	processor []itemproc.ProcessItem) (Scheduler, error) {
	//check the parameters
	if rawMaxDepth == 0 ||
		pageParsers == nil || len(pageParsers) == 0 ||
		processor == nil || len(processor) == 0 {
		return nil, errors.New("The parameters for NewScheduler are illegal.")
	}
	scheduler := &schedulerImpl{}
	atomic.StoreUint32(&(scheduler.status), uint32(SCHEDULER_STATUS_ALLOCATE))
	scheduler.crawMaxDepth = rawMaxDepth
	if err := channelConfig.IsValid(); err != nil {
		atomic.StoreUint32(&(scheduler.status), uint32(SCHEDULER_STATUS_FATAL_ERROR))
		return nil, err
	}
	scheduler.channelConfig = channelConfig

	if err := poolBaseConfig.IsValid(); err != nil {
		atomic.StoreUint32(&(scheduler.status), uint32(SCHEDULER_STATUS_FATAL_ERROR))
		return nil, err
	}
	scheduler.poolBaseConfig = poolBaseConfig

	scheduler.channelManager = util.NewChannelManager(channelConfig)

	dlPool, err := downloader.NewPageDownloaderPoolWithHttpClientGen(poolBaseConfig.PageDownloaderPoolSize(), httpClientGenerator)
	if err != nil {
		return nil, err
	}
	scheduler.dlPool = dlPool

	pageParserPool, err := pageParser.NewPageParserPool(poolBaseConfig.PageParserPoolSize(), func() pageParser.PageParser {
		return pageParser.NewPageParser()
	})
	if err != nil {
		return nil, err
	}
	scheduler.parserPool = pageParserPool

	scheduler.pageParsers = pageParsers
	itemPipeLine, err := itemproc.NewItemPipeline(processor)

	if err != nil {
		return nil, err
	}
	scheduler.itemPipeline = itemPipeLine

	scheduler.stopSign = util.NewStopSign()

	scheduler.reqCache = basic.NewRequestCache(0)
	scheduler.acceptDomain = make(map[string]struct{})
	scheduler.urlMap = make(map[string]bool)

	atomic.StoreUint32(&(scheduler.status), uint32(SCHEDULER_STATUS_READY))

	return scheduler, nil
}

func (sched *schedulerImpl) Start(initRequest *http.Request) error {
	if initRequest == nil {
		return errors.New("The first request cannot be nil.")
	}

	status := atomic.LoadUint32(&sched.status)
	switch status {
	case SCHEDULER_STATUS_ALLOCATE:
		return errors.New("The scheduler is initializing ,please try later.")
	case SCHEDULER_STATUS_RUNNING:
		return errors.New("The scheduler has been started.")
	case SCHEDULER_STATUS_CLOSED:
		return errors.New("The scheduler is closed.")
	case SCHEDULER_STATUS_FATAL_ERROR:
		return errors.New("The scheduler has encountered a fatal error.")
	case SCHEDULER_STATUS_STARTING:
		return errors.New("The scheduler is already starting.")
	}
	atomic.StoreUint32(&(sched.status), uint32(SCHEDULER_STATUS_STARTING))

	sched.startDownloading()
	sched.startPageParsing()
	sched.startItemPipeLine()


	if err := sched.addRequestDomain(initRequest); err != nil {
		return errors.New("Add primary permit domain error.")
	}
	sched.startSchedule(10 * time.Millisecond)

	atomic.StoreUint32(&(sched.status), uint32(SCHEDULER_STATUS_RUNNING))

	sched.reqCache.Put(basic.NewDownloadRequest(0, initRequest, 0))
	return nil
}

func (sched *schedulerImpl) ErrorChan() <-chan error {
	if sched.channelManager.Status() != util.CHANNEL_MANAGER_STATUS_INITIALIZED {
		return nil
	}
	return sched.getErrorChan()
}

func (sched *schedulerImpl) Idle() bool {
	return sched.reqCache.Length() == 0 &&
		len(sched.getReqChan()) == 0 &&
		len(sched.getRespChan()) == 0 &&
		sched.dlPool.Used() == 0 &&
		sched.parserPool.Used() == 0 &&
		sched.itemPipeline.ProcessingNum() == 0
}

func (sched *schedulerImpl) AddPermitDomain(host string) error {
	domain, err := util.GetPrimaryDomain(host)
	if err != nil {
		logs.Error("Add Request domain error: %s\n.", err)
		return err
	}
	if _, ok := sched.acceptDomain[domain]; !ok {
		sched.acceptDomain[domain] = struct{}{}
	}
	return nil
}


func (sched *schedulerImpl) Status() uint {
	return uint(sched.status)
}

func (sched *schedulerImpl) Stop() error {
	if ok := sched.stopSign.SignStop(); !ok {
		return errors.New("The scheduler has been stoped.")
	}
	sched.channelManager.Close()
	sched.reqCache.Close()
	atomic.StoreUint32(&(sched.status), uint32(SCHEDULER_STATUS_CLOSED))

	return nil
}

func (sched *schedulerImpl) Summary() SchedSummary{
	return NewSchedSummary(sched, " ")
}

// 开始下载。
func (sched *schedulerImpl) startDownloading() {
	go func() {
		for {
			req, ok := <-sched.getReqChan()
			if !ok {
				logs.Error("Get request channel Error. \n")
				break
			}
			logs.Debug("scheduler requestchan ")
			go sched.download(req)
		}
	}()
}


func (sched *schedulerImpl) startPageParsing() {
	go func() {
		for {
			resp, ok := <-sched.getRespChan()
			if !ok {
				logs.Error("Get response channel Error.\n")
				break
			}
			go sched.parsePage(sched.pageParsers, resp)
		}
	}()
}

func (sched *schedulerImpl) startItemPipeLine() {
	go func() {
		sched.itemPipeline.SetFastFail(true)
		code := ITEMPIPELINE_CODE
		for item := range sched.getItemChan() {
			go func(item basic.ItemMap) {
				errs := sched.itemPipeline.Send(item)
				if errs != nil {
					for _, err := range errs {
						sched.sendError(err, code)
					}
				}
			}(item)
			defer func() {
				if p := recover(); p != nil {
					errMsg := fmt.Sprintf("Fatal Item Processing Error: %s\n", p)
					logs.Error(errMsg)
				}
			}()
		}
	}()
}

func (sched *schedulerImpl) startSchedule(interval time.Duration) {
	go func() {
		for {
			if sched.stopSign.IsSigned() {
				sched.stopSign.Record(SCHEDULER_CODE)
				return
			}
			if sched.channelManager.Status() == util.CHANNEL_MANAGER_STATUS_INITIALIZED{
				remainder := cap(sched.getReqChan()) - len(sched.getReqChan())
				for ; remainder > 0; remainder-- {

					if sched.stopSign.IsSigned() {
						sched.stopSign.Record(SCHEDULER_CODE)
						return
					}
					newReq := sched.reqCache.Get()
					if newReq==nil {
						continue
					}
					sched.getReqChan() <- newReq
				}
			}
			time.Sleep(interval)
		}
	}()
}

func (sched *schedulerImpl) addRequestDomain(req *http.Request) error {
	err := sched.AddPermitDomain(req.Host)
	if err != nil {
		return err
	}
	return nil
}

func (sched *schedulerImpl) parsePage(parsers []pageParser.ParseResponse, resp *basic.DownloadRespond) {
	defer func() {
		if p := recover(); p != nil {
			logs.Error("Fatal parsing Error: %s\n")
		}
	}()
	pageParser, err := sched.parserPool.Take()

	if err != nil {
		logs.Error("Parser pool Error: %s\n", err)
		sched.sendError(err, SCHEDULER_CODE)
		return
	}
	defer func() {
		err := sched.parserPool.Return(pageParser)
		if err != nil {
			logs.Error("Parser pool Error: %s\n", err)
			sched.sendError(err, SCHEDULER_CODE)
			return
		}
	}()

	code := generateCode(PARSER_CODE, pageParser.Id())

	results, errs := pageParser.ParsePage(parsers, resp)
	if errs != nil {
		for _, err := range errs {
			sched.sendError(err, code)
		}
	}
	if results != nil {
		for _, result := range results {
			if result == nil {
				continue
			}
			switch t := result.(type) {
			case *basic.DownloadRequest:
				sched.sendReqToCache(t, code)
			case basic.ItemMap:
				sched.sendItemMap(t, code)
			default:
				errMsg := fmt.Sprintf("Unsupported data type '%T'! (value=%v)\n", t, t)
				sched.sendError(errors.New(errMsg), code)
			}
		}
	}
}

func (sched *schedulerImpl) download(req *basic.DownloadRequest) {

	defer func() {
		if p := recover(); p != nil {
			logs.Error("Fatal Download Error: %s\n,URL:%s",p)
		}
	}()


	dl, err := sched.dlPool.Take()
	if err != nil {
		logs.Error("Download pool Error: %s\n", err,dl.Id())
		sched.sendError(err, SCHEDULER_CODE)
		return
	}
	defer func() {
		err := sched.dlPool.Return(dl)
		if err != nil {
			logs.Error("Download pool Error: %s\n", err)
			sched.sendError(err, SCHEDULER_CODE)
		}
	}()

	code := generateCode(DOWNLOADER_CODE, dl.Id())

	respond, err := dl.Download(req)

	if err != nil {
		logs.Error("Downloader Error: %s\n", err)
		sched.sendError(err, code)
	}

	if respond != nil {
		sched.sendResp(respond, code)
	}
}

func (sched *schedulerImpl) sendError(err error, code string) bool {
	if err == nil {
		return false
	}
	codePrefix := parseCode(code)[0]
	var errorType basic.ErrorType

	switch codePrefix {
	case DOWNLOADER_CODE:
		errorType = basic.DOWNLOADER_ERROR
	case PARSER_CODE:
		errorType = basic.PAGEPARSER_ERROR
	case ITEMPIPELINE_CODE:
		errorType = basic.ITEM_PROCESSOR_ERROR
	}
	detailErr := basic.NewCrawlerError(errorType, err.Error())
	if sched.stopSign.IsSigned() {
		sched.stopSign.Record(code)
		return false
	}
	go func() {
		sched.getErrorChan() <- detailErr
	}()
	return true
}

func parseCode(code string) []string {
	result := strings.Split(code, "-")
	return result
}

func (sched *schedulerImpl) sendResp(respond *basic.DownloadRespond, code string) bool {
	if sched.stopSign.IsSigned() {
		sched.stopSign.Record(code)
		return false
	}
	if respond == nil {
		logs.Warning("Receive nil response.")
		return false
	}
	sched.getRespChan() <- respond
	return true

}

func (sched *schedulerImpl) sendItemMap(itemMap basic.ItemMap, code string) bool {
	if sched.stopSign.IsSigned() {
		sched.stopSign.Record(code)
		return false
	}
	if itemMap == nil {
		logs.Warning("Receive nil Item map.")
		return false
	}
	sched.getItemChan() <- itemMap
	return true

}

func (sched *schedulerImpl) sendReqToCache(req *basic.DownloadRequest, code string) bool {
	if req == nil || req.HttpReq() == nil {
		logs.Warning("Receive nil request.")
		return false
	}
	reqUrl := req.HttpReq().URL

	if reqUrl == nil {
		logs.Warning("Ignore the request! It's url is is invalid!")
		return false
	}
	if strings.ToLower(reqUrl.Scheme) != "http" {
		if reqUrl.Scheme == "javascript" {
			logs.Debug("Ignore request scheme '%s'.", reqUrl.Scheme)
			return false
		}
		logs.Debug("Find request %s, scheme '%s'.",reqUrl.String(), reqUrl.Scheme)
		//return false
	}
	sched.mapLock.RLock()
	if _, ok := sched.urlMap[reqUrl.String()]; ok {
		logs.Debug("Ignore the request! It's url is repeated. (requestUrl=%s)\n", reqUrl)
		sched.mapLock.RUnlock()
		return false
	}
	sched.mapLock.RUnlock()

	domain, _ := util.GetPrimaryDomain(req.HttpReq().Host)
	if _, ok := sched.acceptDomain[domain]; !ok {
		logs.Debug("Ignore the request! It's host '%s' not in primary domain . (requestUrl=%s)\n",
			req.HttpReq().Host, reqUrl)
		return false
	}
	if req.Depth() > sched.crawMaxDepth {
		logs.Debug("Ignore the request! It's depth %d greater than %d. (requestUrl=%s)\n",
			req.Depth(), sched.crawMaxDepth, reqUrl)
		return false
	}

	if sched.stopSign.IsSigned() {
		sched.stopSign.Record(code)
		return false
	}
	sched.reqCache.Put(req)

	sched.mapLock.Lock()
	sched.urlMap[reqUrl.String()] = true
	sched.mapLock.Unlock()

	return true
}

func generateCode(prefix string, id uint32) string {
	return fmt.Sprintf("%s-%d", prefix, id)
}

func (sched *schedulerImpl) getReqChan() chan *basic.DownloadRequest {
	reqChan, err := sched.channelManager.ReqChan()
	if err != nil {
		panic(err)
	}
	return reqChan
}
func (sched *schedulerImpl) getRespChan() chan *basic.DownloadRespond {
	respChan, err := sched.channelManager.RespChan()
	if err != nil {
		panic(err)
	}
	return respChan
}

func (sched *schedulerImpl) getItemChan() chan basic.ItemMap {
	itemChan, err := sched.channelManager.ItemChan()
	if err != nil {
		panic(err)
	}
	return itemChan
}

func (sched *schedulerImpl) getErrorChan() chan error {
	errChan, err := sched.channelManager.ErrorChan()
	if err != nil {
		panic(err)
	}
	return errChan
}
