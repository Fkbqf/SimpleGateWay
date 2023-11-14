package public

import (
	"FGateWay/golang_common/lib"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"sync/atomic"
	"time"
)

type RedisFlowCountService struct {
	AppID       string
	Interval    time.Duration
	QPS         int64
	Unix        int64
	TickerCount int64
	TotalCount  int64
}

// 创建一个新的 RedisFlowCountService 实例
func NewRedisFlowCountService(appID string, interval time.Duration) *RedisFlowCountService {
	// 初始化 RedisFlowCountService 结构体
	reqCounter := &RedisFlowCountService{
		AppID:    appID,    // 设置应用ID
		Interval: interval, // 设置统计间隔
		QPS:      0,        // 初始每秒请求数为0
		Unix:     0,        // 初始Unix时间戳为0
	}

	// 启动一个新的协程来处理定期统计
	go func() {
		// 如果协程中有任何错误，捕获并打印它
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()

		// 创建一个新的定时器，每隔interval时间就触发一次
		ticker := time.NewTicker(interval)
		for {
			<-ticker.C // 等待下一个定时器触发

			tickerCount := atomic.LoadInt64(&reqCounter.TickerCount)
			atomic.StoreInt64(&reqCounter.TickerCount, 0)

			// 获取当前时间
			currentTime := time.Now()
			// 生成当前日期和小时的键
			dayKey := reqCounter.GetDayKey(currentTime)
			hourKey := reqCounter.GetHourKey(currentTime)

			// 使用Redis的pipeline更新日和小时的请求计数
			if err := RedisConfPipline(func(c redis.Conn) {
				c.Send("INCRBY", dayKey, tickerCount)  // 增加当日计数
				c.Send("EXPIRE", dayKey, 86400*2)      // 设置当日计数的过期时间为2天
				c.Send("INCRBY", hourKey, tickerCount) // 增加当前小时的计数
				c.Send("EXPIRE", hourKey, 86400*2)     // 设置当前小时的过期时间为2天
			}); err != nil {
				// 如果有错误，打印并继续下一个周期
				fmt.Println("RedisConfPipline err", err)
				continue
			}

			// 从Redis获取当天的总请求计数
			totalCount, err := reqCounter.GetDayData(currentTime)
			if err != nil {
				// 如果有错误，打印并继续下一个周期
				fmt.Println("reqCounter.GetDayData err", err)
				continue
			}
			// 获取当前的Unix时间戳
			nowUnix := time.Now().Unix()
			if reqCounter.Unix == 0 {
				reqCounter.Unix = time.Now().Unix()
				continue
			}
			// 计算新的QPS值
			tickerCount = totalCount - reqCounter.TotalCount
			if nowUnix > reqCounter.Unix {
				reqCounter.TotalCount = totalCount
				reqCounter.QPS = tickerCount / (nowUnix - reqCounter.Unix)
				reqCounter.Unix = time.Now().Unix()
			}
		}
	}()

	// 返回初始化的RedisFlowCountService实例
	return reqCounter
}

func (o *RedisFlowCountService) GetDayKey(t time.Time) string {
	dayStr := t.In(lib.TimeLocation).Format("20060102")
	return fmt.Sprintf("%s_%s_%s", RedisFlowDayKey, dayStr, o.AppID)
}

func (o *RedisFlowCountService) GetHourKey(t time.Time) string {
	hourStr := t.In(lib.TimeLocation).Format("2006010215")
	return fmt.Sprintf("%s_%s_%s", RedisFlowHourKey, hourStr, o.AppID)
}

func (o *RedisFlowCountService) GetHourData(t time.Time) (int64, error) {
	return redis.Int64(RedisConfDo("GET", o.GetHourKey(t)))
}

func (o *RedisFlowCountService) GetDayData(t time.Time) (int64, error) {
	return redis.Int64(RedisConfDo("GET", o.GetDayKey(t)))
}

// 原子增加
func (o *RedisFlowCountService) Increase() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		atomic.AddInt64(&o.TickerCount, 1)
	}()
}
