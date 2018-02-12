package basic

import (
	"sync"
	"errors"
	"fmt"
)

type CacheStatus int

const (
	STATUS_RUNNING CacheStatus = iota
	STATUS_CLOSED
	STATUS_UNKNOWN
)

var statusMsg = map[CacheStatus]string{
	STATUS_RUNNING: "RUNNING",
	STATUS_CLOSED:    "CLOSED",
	STATUS_UNKNOWN: "UNKNOWN",
}

type RequestCache interface {
	Put(req *DownloadRequest) error
	Get() *DownloadRequest
	Capacity() int
	Length() int
	Close()
	Open ()
	Summary() string
}

type requestCacheImpl struct {
	sync.Mutex
	cache  []*DownloadRequest
	status CacheStatus
}

func NewRequestCache(capacity uint) RequestCache {
	return &requestCacheImpl{cache: make([]*DownloadRequest, capacity)}
}

func (rci *requestCacheImpl) Put(req *DownloadRequest) error {
	if req==nil {
		return errors.New("The request can not be nil.")
	}
	if rci.status==STATUS_CLOSED{
		return errors.New("The cache has been closed.")
	}
	rci.Lock()
	defer rci.Unlock()
	rci.cache=append(rci.cache,req)
	return nil
}

func (rci *requestCacheImpl) Get( ) *DownloadRequest {
	if rci.status==STATUS_CLOSED || rci.Length()==0 {
		return nil
	}
	rci.Lock()
	defer rci.Unlock()
	req:=rci.cache[0]
	rci.cache=rci.cache[1:]
	return req
}

func (rci *requestCacheImpl) Length() int{
	return len(rci.cache)
}

func (rci *requestCacheImpl) Capacity() int {
	return cap(rci.cache)
}


func (rci *requestCacheImpl) Close() {
	rci.status=STATUS_CLOSED
}

func (rci *requestCacheImpl) Open() {
	rci.status=STATUS_RUNNING
}


// 摘要信息模板。
var summaryTemplate = "status: %s, " + "length: %d, " + "capacity: %d"

func (rci *requestCacheImpl) Summary() string {
	summary := fmt.Sprintf(summaryTemplate,
		statusMsg[rci.status],
		rci.Length(),
		rci.Capacity())
	return summary
}
