package scheduler

import (
	//"bytes"
	"fmt"
	"chaoshen.com/crawlergo/crawler/basic"
)

// 调度器摘要信息的接口类型。
type SchedSummary interface {
	String() string               // 获得摘要信息的一般表示。
	Detail() string               // 获取摘要信息的详细表示。
	Same(other SchedSummary) bool // 判断是否与另一份摘要信息相同。
}

// 创建调度器摘要信息。
func NewSchedSummary(sched *schedulerImpl, prefix string) SchedSummary {
	if sched == nil {
		return nil
	}
	//urlCount := len(sched.urlMap)
	var urlDetail string
	/*
	if urlCount > 0 {
		var buffer bytes.Buffer
		buffer.WriteByte('\n')
		for k, _ := range sched.urlMap {
			buffer.WriteString(prefix)
			buffer.WriteString(prefix)
			buffer.WriteString(k)
			buffer.WriteByte('\n')
		}
		urlDetail = buffer.String()
	} else {
	*/
		urlDetail = "\n"
	//}
	return &schedSummaryImpl{
		prefix:              prefix,
		status:              sched.status,
		channelConfig:       sched.channelConfig,
		poolBaseConfig:      sched.poolBaseConfig,
		crawMaxDepth:         sched.crawMaxDepth,
		chanManSummary:      sched.channelManager.Summary(),
		reqCacheSummary:     sched.reqCache.Summary(),
		dlPoolLen:           sched.dlPool.Used(),
		dlPoolCap:           sched.dlPool.Total(),
		analyzerPoolLen:     sched.parserPool.Used(),
		analyzerPoolCap:     sched.parserPool.Total(),
		itemPipelineSummary: sched.itemPipeline.Summary(),
		urlCount:            0,
		urlDetail:           urlDetail,
		stopSignSummary:     sched.stopSign.Summary(),
	}
}

// 调度器摘要信息的实现类型。
type schedSummaryImpl struct {
	prefix              string            // 前缀。
	status              uint32            // 运行标记。
	channelConfig       basic.ChannelConfig  // 通道参数的容器。
	poolBaseConfig      basic.PoolBaseConfig // 池基本参数的容器。
	crawMaxDepth        uint32            // 爬取的最大深度。
	chanManSummary      string            // 通道管理器的摘要信息。
	reqCacheSummary     string            // 请求缓存的摘要信息。
	dlPoolLen           uint32            // 网页下载器池的长度。
	dlPoolCap           uint32            // 网页下载器池的容量。
	analyzerPoolLen     uint32            // 分析器池的长度。
	analyzerPoolCap     uint32            // 分析器池的容量。
	itemPipelineSummary string            // 条目处理管道的摘要信息。
	urlCount            int               // 已请求的URL的计数。
	urlDetail           string            // 已请求的URL的详细信息。
	stopSignSummary     string            // 停止信号的摘要信息。
}

func (ss *schedSummaryImpl) String() string {
	return ss.getSummary(false)
}

func (ss *schedSummaryImpl) Detail() string {
	return ss.getSummary(true)
}

// 获取摘要信息。
func (ss *schedSummaryImpl) getSummary(detail bool) string {
	prefix := ss.prefix
	template := prefix + "Running: %v \n" +
		prefix + "Channel config: %s \n" +
		prefix + "Pool base config: %s \n" +
		prefix + "Crawl depth: %d \n" +
		prefix + "Channels manager: %s \n" +
		prefix + "Request cache: %s\n" +
		prefix + "Downloader pool: %d/%d\n" +
		prefix + "parser pool: %d/%d\n" +
		prefix + "Item pipeline: %s\n" +
		prefix + "Urls(%d): %s" +
		prefix + "Stop sign: %s\n"
	return fmt.Sprintf(template,
		func() bool {
			return ss.status == SCHEDULER_STATUS_RUNNING
		}(),
		ss.channelConfig.Summary(),
		ss.poolBaseConfig.Summary(),
		ss.crawMaxDepth,
		ss.chanManSummary,
		ss.reqCacheSummary,
		ss.dlPoolLen, ss.dlPoolCap,
		ss.analyzerPoolLen, ss.analyzerPoolCap,
		ss.itemPipelineSummary,
		ss.urlCount,
		func() string {
			if detail {
				return ss.urlDetail
			} else {
				return "<concealed>\n"
			}
		}(),
		ss.stopSignSummary)
}

func (ss *schedSummaryImpl) Same(other SchedSummary) bool {
	if other == nil {
		return false
	}
	otherSs, ok := interface{}(other).(*schedSummaryImpl)
	if !ok {
		return false
	}
	if ss.status != otherSs.status ||
		ss.crawMaxDepth != otherSs.crawMaxDepth ||
		ss.dlPoolLen != otherSs.dlPoolLen ||
		ss.dlPoolCap != otherSs.dlPoolCap ||
		ss.analyzerPoolLen != otherSs.analyzerPoolLen ||
		ss.analyzerPoolCap != otherSs.analyzerPoolCap ||
		ss.urlCount != otherSs.urlCount ||
		ss.stopSignSummary != otherSs.stopSignSummary ||
		ss.reqCacheSummary != otherSs.reqCacheSummary ||
		ss.poolBaseConfig.Summary() != otherSs.poolBaseConfig.Summary() ||
		ss.channelConfig.Summary() != otherSs.channelConfig.Summary() ||
		ss.itemPipelineSummary != otherSs.itemPipelineSummary ||
		ss.chanManSummary != otherSs.chanManSummary {
		return false
	} else {
		return true
	}
}
