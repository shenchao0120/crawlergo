package basic

import (
	"errors"
	"fmt"
)

type Config interface {
	IsValid() error
	Summary() string
}

type ChannelConfig struct {
	reqChanLen   uint   // 请求通道的长度。
	respChanLen  uint   // 响应通道的长度。
	itemChanLen  uint   // 条目通道的长度。
	errorChanLen uint   // 错误通道的长度。
	summary      string // 描述。
}

func NewChannelConfig(reqChanLen uint, respChanLen uint, itemChanLen uint, errorChanLen uint) ChannelConfig {
	return ChannelConfig{
		reqChanLen:   reqChanLen,
		respChanLen:  respChanLen,
		itemChanLen:  itemChanLen,
		errorChanLen: errorChanLen,
	}
}

func (config *ChannelConfig) IsValid() error {
	if config.reqChanLen == 0 {
		return errors.New("The request channel max length (capacity) can not be 0!\n")
	}
	if config.respChanLen == 0 {
		return errors.New("The response channel max length (capacity) can not be 0!\n")
	}
	if config.itemChanLen == 0 {
		return errors.New("The item channel max length (capacity) can not be 0!\n")
	}
	if config.errorChanLen == 0 {
		return errors.New("The error channel max length (capacity) can not be 0!\n")
	}
	return nil
}

var channelArgsTemplate = "{ reqChanLen: %d, respChanLen: %d," +
	" itemChanLen: %d, errorChanLen: %d }"

func (config *ChannelConfig) Summary() string {
	if config.summary == "" {
		config.summary =
			fmt.Sprintf(channelArgsTemplate,
				config.reqChanLen,
				config.respChanLen,
				config.itemChanLen,
				config.errorChanLen)
	}
	return config.summary
}

func (config *ChannelConfig) ReqChanLen() uint {
	return config.respChanLen
}

func (config *ChannelConfig) RespChanLen() uint {
	return config.respChanLen
}

func (config *ChannelConfig) ItemChanLen() uint {
	return config.itemChanLen
}

func (config *ChannelConfig) ErrorChanLen() uint {
	return config.errorChanLen
}

// 池基本参数容器的描述模板。
var poolBaseConfigTemplate string = "{ pageDownloaderPoolSize: %d," +
	" PageParserPoolSize: %d }"

// 池基本参数的容器。
type PoolBaseConfig struct {
	pageDownloaderPoolSize uint32 // 网页下载器池的尺寸。
	pageParserPoolSize     uint32 //
	summary                string // 描述。
}

func NewPoolBaseConfig(downloaderPoolSize uint32,parserPoolSize uint32)PoolBaseConfig{
	return PoolBaseConfig{pageDownloaderPoolSize:downloaderPoolSize,pageParserPoolSize:parserPoolSize}
}

func (config *PoolBaseConfig) IsValid() error {
	if config.pageDownloaderPoolSize == 0 {
		return errors.New("The page downloader pool size can not be 0!\n")
	}
	if config.pageParserPoolSize == 0 {
		return errors.New("The page parser pool size can not be 0!\n")
	}
	return nil
}

func (config *PoolBaseConfig) Summary() string {
	if config.summary == "" {
		config.summary =
			fmt.Sprintf(poolBaseConfigTemplate,
				config.pageDownloaderPoolSize,
				config.pageParserPoolSize)
	}
	return config.summary
}

// 获得网页下载器池的尺寸。
func (config *PoolBaseConfig) PageDownloaderPoolSize() uint32 {
	return config.pageDownloaderPoolSize
}

// 获得分析器池的尺寸。
func (config *PoolBaseConfig) PageParserPoolSize() uint32 {
	return config.pageParserPoolSize
}
