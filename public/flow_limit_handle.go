package public

import (
	"golang.org/x/time/rate" // 导入外部包，用于令牌桶算法限流
	"sync"                   // 导入同步包，提供了互斥锁等并发原语
)

var FlowLimiterHandler *FlowLimiter // 定义全局变量FlowLimiterHandler，用于存储流控处理器的单例实例

// FlowLimiter结构体，用于存储流控信息
type FlowLimiter struct {
	FlowLmiterMap   map[string]*FlowLimiterItem // 使用map存储每个服务的流控信息，key为服务名，value为流控项
	FlowLmiterSlice []*FlowLimiterItem          // 使用切片存储流控项
	Locker          sync.RWMutex                // 定义一个读写锁，用于保护流控信息的并发访问
}

// FlowLimiterItem结构体，用于存储每个服务的流控信息
type FlowLimiterItem struct {
	ServiceName string        // 服务名称
	Limter      *rate.Limiter // 每个服务的限流器
}

// NewFlowLimiter函数，用于创建一个新的FlowLimiter实例
func NewFlowLimiter() *FlowLimiter {
	return &FlowLimiter{
		FlowLmiterMap:   map[string]*FlowLimiterItem{},
		FlowLmiterSlice: []*FlowLimiterItem{},
		Locker:          sync.RWMutex{},
	}
}

// init函数，当public包被其他代码导入时自动执行
func init() {
	FlowLimiterHandler = NewFlowLimiter() // 初始化FlowLimiterHandler为一个新的FlowLimiter实例
}

// GetLimiter函数，用于获取指定服务的限流器，如果不存在则创建一个新的限流器
func (counter *FlowLimiter) GetLimiter(serverName string, qps float64) (*rate.Limiter, error) {
	// 遍历FlowLmiterSlice，检查是否已经存在指定服务的限流器
	for _, item := range counter.FlowLmiterSlice {
		if item.ServiceName == serverName {
			return item.Limter, nil // 如果找到，直接返回这个限流器
		}
	}

	// 如果没有找到，创建一个新的限流器，令牌桶算法，桶的大小是qps的3倍
	newLimiter := rate.NewLimiter(rate.Limit(qps), int(qps*3))
	// 创建一个新的FlowLimiterItem
	item := &FlowLimiterItem{
		ServiceName: serverName,
		Limter:      newLimiter,
	}
	// 将新的FlowLimiterItem添加到FlowLmiterSlice中
	counter.FlowLmiterSlice = append(counter.FlowLmiterSlice, item)
	counter.Locker.Lock()         // 加写锁，保护FlowLmiterMap的并发修改
	defer counter.Locker.Unlock() // 解锁
	// 将新的FlowLimiterItem添加到FlowLmiterMap中
	counter.FlowLmiterMap[serverName] = item
	return newLimiter, nil // 返回新创建的限流器
}
