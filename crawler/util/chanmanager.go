package util

import (
	. "chaoshen.com/crawlergo/crawler/basic"
	"sync"
	"errors"
	"github.com/astaxie/beego/logs"
	"fmt"
)

type ChannelManagerStatus int

const (
	CHANNEL_MANAGER_STATUS_UNINITIALIZED ChannelManagerStatus = iota // 未初始化状态。
	CHANNEL_MANAGER_STATUS_INITIALIZED    							 // 已初始化状态。
	CHANNEL_MANAGER_STATUS_CLOSED         							 // 已关闭状态。
)


var statusNameMap=map[ChannelManagerStatus]string{
	CHANNEL_MANAGER_STATUS_UNINITIALIZED:"uninitialized",
	CHANNEL_MANAGER_STATUS_INITIALIZED:"initialized",
	CHANNEL_MANAGER_STATUS_CLOSED:"closed",
}

type ChannelManager interface {
	Init(config ChannelConfig ,reset bool) error
	Close() error
	ReqChan () (chan *DownloadRequest,error)
	RespChan () (chan *DownloadRespond,error)
	ItemChan () (chan ItemMap,error)
	ErrorChan () (chan error,error)
	Status () ChannelManagerStatus
	Summary() string
}



type channelManagerImpl struct {
	sync.RWMutex
	config ChannelConfig
	reqChan chan *DownloadRequest
	respChan chan *DownloadRespond
	itemChan chan ItemMap
	errorChan chan error
	status ChannelManagerStatus
}

func NewChannelManager(config ChannelConfig) ChannelManager{
	chanManager:=&channelManagerImpl{}
	chanManager.Init(config,true)
	return chanManager
}

func (cmi *channelManagerImpl)Init(config ChannelConfig ,reset bool) error{
	if err:=config.IsValid() ; err!=nil {
		return err
	}
	cmi.Lock()
	defer cmi.Unlock()
	if reset==false && cmi.status==CHANNEL_MANAGER_STATUS_INITIALIZED {
		return errors.New("The channel manager has been initialized.")
	}
	cmi.config=config
	cmi.reqChan=make(chan *DownloadRequest,config.ReqChanLen())
	cmi.respChan=make(chan *DownloadRespond,config.RespChanLen())
	cmi.itemChan=make(chan ItemMap,config.ItemChanLen())
	cmi.errorChan=make(chan error,config.ErrorChanLen())
	cmi.status=CHANNEL_MANAGER_STATUS_INITIALIZED
	return nil
}

func (cmi *channelManagerImpl)Close()error{
	cmi.Lock()
	defer cmi.Unlock()
	if cmi.status==CHANNEL_MANAGER_STATUS_UNINITIALIZED {
		return errors.New("Cannot close a uninitialized channel manager.")
	}
	if cmi.status==CHANNEL_MANAGER_STATUS_CLOSED {
		logs.Info("The channel manager has benn closed.")
		return nil
	}
	close(cmi.reqChan)
	close(cmi.respChan)
	close(cmi.itemChan)
	close(cmi.errorChan)
	cmi.status=CHANNEL_MANAGER_STATUS_CLOSED
	return nil
}

func (cmi *channelManagerImpl)ReqChan() (chan *DownloadRequest,error){
	cmi.RLock()
	defer cmi.RUnlock()
	if err:=cmi.checkStatus() ;err!=nil{
		return nil,err
	}
	return cmi.reqChan,nil
}

func (cmi *channelManagerImpl)RespChan() (chan *DownloadRespond,error){
	cmi.RLock()
	defer cmi.RUnlock()
	if err:=cmi.checkStatus() ;err!=nil{
		return nil,err
	}
	return cmi.respChan,nil
}

func (cmi *channelManagerImpl)ItemChan() (chan ItemMap,error){
	cmi.RLock()
	defer cmi.RUnlock()
	if err:=cmi.checkStatus() ;err!=nil{
		return nil,err
	}
	return cmi.itemChan,nil
}

func (cmi *channelManagerImpl)ErrorChan() (chan error,error){
	cmi.RLock()
	defer cmi.RUnlock()
	if err:=cmi.checkStatus() ;err!=nil{
		return nil,err
	}
	return cmi.errorChan,nil
}

func (cmi *channelManagerImpl)Status () ChannelManagerStatus{
	return cmi.status
}


func (cmi *channelManagerImpl) checkStatus() error {
	if cmi.status == CHANNEL_MANAGER_STATUS_INITIALIZED {
		return nil
	}
	statusName, ok := statusNameMap[cmi.status]
	if !ok {
		statusName = fmt.Sprintf("%d", cmi.status)
	}
	errMsg :=
		fmt.Sprintf("The undesirable status of channel manager: %s!\n",
			statusName)
	return errors.New(errMsg)
}

var chanmanSummaryTemplate = "status: %s, " +
	"requestChannel: %d/%d, " +
	"responseChannel: %d/%d, " +
	"itemChannel: %d/%d, " +
	"errorChannel: %d/%d"

func (cmi *channelManagerImpl)Summary() string{
	summary := fmt.Sprintf(chanmanSummaryTemplate,
		statusNameMap[cmi.status],
		len(cmi.reqChan), cap(cmi.reqChan),
		len(cmi.respChan), cap(cmi.respChan),
		len(cmi.itemChan), cap(cmi.itemChan),
		len(cmi.errorChan), cap(cmi.errorChan))
	return summary
}
