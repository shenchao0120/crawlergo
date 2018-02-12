package downloader

import (
	"chaoshen.com/crawlergo/crawler/basic"
	"reflect"
	"net/http"
	"github.com/astaxie/beego/logs"
	"errors"
)


type PageDownloaderPool interface {
	Take() (PageDownloader, error)
	Return(downloader PageDownloader) error
	Total() uint32
	Used() uint32
}

type GenPageDownloader func() PageDownloader

type GenHttpClient func() *http.Client

type PageDownloaderPoolImpl struct {
	pool           basic.Pool
	downloaderType reflect.Type
}

func NewPageDownloaderPoolWithGen(total uint32, gen GenPageDownloader) (PageDownloaderPool, error) {
	entityType := reflect.TypeOf(gen())
	genEntity := func() basic.Entity {
		return gen()
	}
	pool, err := basic.NewPool(total, entityType, genEntity)
	if err != nil {
		logs.Error("New pool error ", err)
		return nil, err
	}
	dlpool:=&PageDownloaderPoolImpl{pool:pool,downloaderType:entityType}
	return dlpool,nil
}

func NewPageDownloaderPoolWithHttpClientGen(total uint32, gen GenHttpClient) (PageDownloaderPool, error) {
	if gen==nil {
		gen = func() *http.Client {
			return &http.Client{}
		}
	}
	genEntity := func() basic.Entity {
		return NewPageDownloader(gen())
	}
	entityType:=reflect.TypeOf(genEntity())
	pool, err := basic.NewPool(total, entityType, genEntity)
	if err != nil {
		logs.Error("New pool error ", err)
		return nil, err
	}
	dlPool:=&PageDownloaderPoolImpl{pool:pool,downloaderType:entityType}
	return dlPool,nil
}

func NewPageDownloaderPool(total uint32) (PageDownloaderPool, error) {
	return NewPageDownloaderPoolWithHttpClientGen(total,nil)
}

func (dlPool *PageDownloaderPoolImpl) Take() (PageDownloader,error){
	entity,err:=dlPool.pool.Take()
	if err!=nil {
		return nil,err
	}
	dl,ok:=entity.(PageDownloader)
	if !ok{
		return nil,errors.New("The pool entity is not page downloader.")
	}
	return dl,nil
}

func (dlPool *PageDownloaderPoolImpl) Return(pdl PageDownloader) error{
		return dlPool.pool.Return(pdl)
}

func (dlPool *PageDownloaderPoolImpl) Total() uint32{
	return dlPool.pool.Total()
}

func (dlPool *PageDownloaderPoolImpl) Used() uint32{
	return dlPool.pool.Used()
}
