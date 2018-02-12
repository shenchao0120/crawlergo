package pageParser

import (
	"chaoshen.com/crawlergo/crawler/basic"
	"reflect"
	"errors"
)


type PageParserPool interface {
	Take() (PageParser, error)
	Return(downloader PageParser) error
	Total() uint32
	Used() uint32
}

type GenPageParser func() PageParser


type pageParserPoolImpl struct {
	pool            basic.Pool
	parserType		reflect.Type
}



func NewPageParserPool(total uint32,gen GenPageParser) (PageParserPool, error) {
	if gen==nil{
		return nil ,errors.New("The genPageParser function cannot be nil")
	}
	eType:=reflect.TypeOf(gen())
	genEntity:= func()basic.Entity {
		return gen()
	}
	pool,err:=basic.NewPool(total,eType,genEntity)
	if err!=nil{
		return nil,err
	}
	return &pageParserPoolImpl{pool:pool,parserType:eType},nil

}

func (ppPool *pageParserPoolImpl) Take() (PageParser,error){
	entity,err:=ppPool.pool.Take()
	if err!=nil {
		return nil,err
	}
	dl,ok:=entity.(PageParser)
	if !ok{
		return nil,errors.New("The pool entity is not page downloader.")
	}
	return dl,nil
}

func (ppPool *pageParserPoolImpl) Return(pdl PageParser) error{
	return ppPool.pool.Return(pdl)
}

func (ppPool *pageParserPoolImpl) Total() uint32{
	return ppPool.pool.Total()
}

func (ppPool *pageParserPoolImpl) Used() uint32{
	return ppPool.pool.Used()
}
