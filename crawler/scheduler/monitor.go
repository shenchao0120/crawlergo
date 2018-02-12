package scheduler

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"runtime"
	"time"
)

var summaryForMonitoring = "Monitor - Collected information[%d]:\n" +
	"  Goroutine number: %d\n" +
	"  Scheduler:\n%s" +
	"  Escaped time: %s\n"

var msgStopScheduler = "Stop scheduler...%s."

func Monitoring(
	scheduler Scheduler,
	intervalNs time.Duration,
	detailSummary bool) <-chan bool {
	if scheduler == nil { // 调度器不能不可用！
		panic(errors.New("The scheduler is invalid!"))
	}

	flagChan := make(chan bool)
	// 防止过小的参数值对爬取流程的影响
	if intervalNs < time.Millisecond {
		intervalNs = time.Millisecond
	}
	// 监控停止通知器
	stopNotifier := make(chan byte, 1)
	// 接收和报告错误
	reportError(scheduler, stopNotifier)
	// 记录摘要信息
	recordSummary(scheduler, detailSummary, stopNotifier)
	// 检查空闲状态
	checkStatus(scheduler,
		flagChan,
		intervalNs,
		stopNotifier)
	return flagChan
}

func checkStatus(
	scheduler Scheduler,
	flagChan chan bool,
	intervalNs time.Duration,
	stopNotifier chan<- byte) {
	go func() {
		defer func() {
			stopNotifier <- 1
			stopNotifier <- 2
			flagChan <- true
		}()
		// 等待调度器开启
		waitForSchedulerStart(scheduler)
		for {
			// 检查调度器的空闲状态
			if scheduler.Status() == SCHEDULER_STATUS_CLOSED {
				return
			}
			if scheduler.Idle() {
				var result string
				if err := scheduler.Stop(); err == nil {
					result = "success"
				} else {
					result = "failing"
				}
				msg := fmt.Sprintf(msgStopScheduler, result)
				logs.Critical(msg)
			}
			time.Sleep(intervalNs)
		}
	}()
}

func waitForSchedulerStart(scheduler Scheduler) {
	for scheduler.Status() != SCHEDULER_STATUS_RUNNING {
		time.Sleep(time.Millisecond)
	}
}

func reportError(scheduler Scheduler, stopNotifier <-chan byte) {
	go func() {
		waitForSchedulerStart(scheduler)
		for {
			select {
			case <-stopNotifier:
				return
			default:
			}
			errorChan := scheduler.ErrorChan()
			if errorChan == nil {
				return
			}
			errSend := <-errorChan
			if errSend != nil {
				errMsg := fmt.Sprintf("Error (received from error channel): %s", errSend)
				logs.Critical(errMsg)
			}
			time.Sleep(1 * time.Millisecond)
		}
	}()
}

func recordSummary(scheduler Scheduler, detailSummary bool, stopNotifier <-chan byte) {
	go func() {
		waitForSchedulerStart(scheduler)
		var prevSummary SchedSummary
		var prevNumGoroutine int
		var recordCount uint64 = 1
		startTime := time.Now()

		for {
			select {
			case <-stopNotifier:
				return
			default:
			}
			currNumGoroutine := runtime.NumGoroutine()
			currSchedSummary := scheduler.Summary()
			if currNumGoroutine != prevNumGoroutine || !prevSummary.Same(currSchedSummary) {
				schedSummaryStr := func() string {
					if detailSummary == true {
						return currSchedSummary.Detail()
					} else {
						return currSchedSummary.String()
					}
				}()
				info := fmt.Sprintf(summaryForMonitoring,
					recordCount,
					currNumGoroutine,
					schedSummaryStr,
					time.Since(startTime).String(),
				)
				logs.Info("Monitor record:%s", info)
				recordCount++
				prevNumGoroutine = currNumGoroutine
				prevSummary = currSchedSummary
			}
			time.Sleep(1 * time.Second)
		}
	}()
}
