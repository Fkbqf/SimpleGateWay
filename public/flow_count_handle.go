package public

import (
	"sync"
	"time"
)

var FlowCounterHandler *FlowCounter

type FlowCounter struct {
	RedisFlowCountMap   map[string]*RedisFlowCountService
	RedisFlowCountSlice []*RedisFlowCountService
	Locker              sync.RWMutex
}

func (c *FlowCounter) GetCounter(serverName string) (*RedisFlowCountService, error) {
	for _, item := range c.RedisFlowCountSlice {
		if item.AppID == serverName {
			return item, nil
		}
	}

	newCounter := NewRedisFlowCountService(serverName, 1*time.Second)
	c.RedisFlowCountSlice = append(c.RedisFlowCountSlice, newCounter)
	c.Locker.Lock()
	defer c.Locker.Unlock()
	c.RedisFlowCountMap[serverName] = newCounter
	return newCounter, nil
}

func NewFlowCounter() *FlowCounter {
	return &FlowCounter{
		RedisFlowCountMap:   map[string]*RedisFlowCountService{},
		RedisFlowCountSlice: []*RedisFlowCountService{},
		Locker:              sync.RWMutex{},
	}
}

func init() {
	FlowCounterHandler = NewFlowCounter()
}
