package itemproc

import (
	"chaoshen.com/crawlergo/crawler/basic"
	"qiniupkg.com/x/errors.v7"
	"fmt"
	"sync/atomic"
)

type ItemPipeline interface {
	Send(itemsMap basic.ItemMap) []error
	FastFail() bool
	SetFastFail(fastFail bool)
	SentNum() uint64
	AcceptedNum() uint64
	ProcessedNum() uint64
	ProcessingNum() uint64
	Summary() string
}

type itemPipelineImpl struct {
	itemProcessors []ProcessItem // 条目处理器的列表。
	failFast       bool          // 表示处理是否需要快速失败的标志位。
	sentNum        uint64        // 已被发送的条目的数量。
	acceptedNum    uint64        // 已被接受的条目的数量。
	processedNum   uint64        // 已被处理的条目的数量。
	processingNum  uint64        // 正在被处理的条目的数量。
}

func NewItemPipeline(itemProcessors []ProcessItem) (ItemPipeline,error){
	if itemProcessors==nil {
		return nil,errors.New("Invalid item processor list!")
	}
	for i,processor:=range itemProcessors {
		if processor== nil {
			return nil,errors.New(fmt.Sprintf("Invalid item processor[%d]!\n", i))
		}
	}

	return &itemPipelineImpl{itemProcessors:itemProcessors},nil
}


func (ppi * itemPipelineImpl)Send(itemsMap basic.ItemMap)[]error{
	atomic.AddUint64(&ppi.sentNum,1)
	atomic.AddUint64(&ppi.processingNum,1)
	defer atomic.AddUint64(&ppi.processingNum, ^uint64(0))

	errs:=make([]error,0)

	if itemsMap==nil {
		errs:=append(errs,errors.New("The input item is nil"))
		return errs
	}
	atomic.AddUint64(&ppi.acceptedNum,1)
	var currentItem=itemsMap

	for _,processor:=range ppi.itemProcessors{
		tempItem,err:=processor(currentItem)
		if err!=nil {
			errs = append(errs, err)
			if ppi.failFast == true {
				break
			}
		}
		if tempItem!=nil{
			currentItem=tempItem
		}
	}
	atomic.AddUint64(&ppi.processedNum,1)
	return errs
}

func (ppi *itemPipelineImpl) FastFail() bool{
	return ppi.failFast
}

func (ppi *itemPipelineImpl) SetFastFail(fastFail bool){
	ppi.failFast=fastFail
}

func (ppi *itemPipelineImpl) SentNum() uint64{
	return atomic.LoadUint64(&ppi.sentNum)
}

func (ppi *itemPipelineImpl) AcceptedNum() uint64{
	return atomic.LoadUint64(&ppi.acceptedNum)
}

func (ppi *itemPipelineImpl) ProcessedNum() uint64{
	return atomic.LoadUint64(&ppi.processedNum)
}

func (ppi *itemPipelineImpl) ProcessingNum() uint64{
	return atomic.LoadUint64(&ppi.processingNum)
}


var summaryTemplate = "FailFast: %v, processorNumber: %d," +
	" sent: %d, accepted: %d, processed: %d, processingNumber: %d"

func (ppi *itemPipelineImpl) Summary() string {
	summary := fmt.Sprintf(summaryTemplate,
		ppi.failFast, len(ppi.itemProcessors),ppi.sentNum,ppi.acceptedNum,ppi.processedNum,ppi.processingNum)
	return summary
}





