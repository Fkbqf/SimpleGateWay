package dao

import (
	"FGateWay/dto"
	"FGateWay/golang_common/lib"
	"FGateWay/public"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http/httptest"
	"strings"
	"sync"
)

type ServiceDetail struct {
	Info          *ServiceInfo   `json:"info" description:"基本信息"`
	HTTPRule      *HttpRule      `json:"http_rule" description:"http_rule"`
	TCPRule       *TcpRule       `json:"tcp_rule" description:"tcp_rule"`
	GRPCRule      *GrpcRule      `json:"grpc_rule" description:"grpc_rule"`
	LoadBalance   *LoadBalance   `json:"load_balance" description:"load_balance"`
	AccessControl *AccessControl `json:"access_control" description:"access_control"`
}

var ServiceManagerHandler *ServiceManager

func init() {
	ServiceManagerHandler = NewServiceManager()
}

type ServiceManager struct {
	ServiceMap   map[string]*ServiceDetail
	ServiceSlice []*ServiceDetail
	Locker       sync.RWMutex
	init         sync.Once
	err          error
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		ServiceMap:   map[string]*ServiceDetail{},
		ServiceSlice: []*ServiceDetail{},
		Locker:       sync.RWMutex{},
		init:         sync.Once{},
	}
}
func (t *ServiceInfo) GroupByLoadType(c *gin.Context, tx *gorm.DB) ([]dto.DashServiceStatItemOutput, error) {
	list := []dto.DashServiceStatItemOutput{}
	query := tx.SetCtx(public.GetGinTraceContext(c))
	if err := query.Table(t.TableName()).Where("is_delete=0").Select("load_type, count(*) as value").Group("load_type").Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
func (s *ServiceManager) GetTcpServiceList() []*ServiceDetail {
	list := []*ServiceDetail{}
	for _, serverItem := range s.ServiceSlice {
		tempItem := serverItem
		if tempItem.Info.LoadType == public.LoadTypeTCP {
			list = append(list, tempItem)
		}
	}
	return list
}

func (s *ServiceManager) GetGrpcServiceList() []*ServiceDetail {
	list := []*ServiceDetail{}
	for _, serverItem := range s.ServiceSlice {
		tempItem := serverItem
		if tempItem.Info.LoadType == public.LoadTypeGRPC {
			list = append(list, tempItem)
		}
	}
	return list
}

// 接入方式中间件
func (s *ServiceManager) HTTPAccessMode(c *gin.Context) (*ServiceDetail, error) {
	//1、前缀匹配 /abc ==> serviceSlice.rule
	//2、域名匹配 www.test.com ==> serviceSlice.rule
	//host c.Request.Host
	//path c.Request.URL.Path
	host := c.Request.Host                  //        www.test.com:8080
	host = host[0:strings.Index(host, ":")] //www.test.com
	path := c.Request.URL.Path              // /abc/get

	for _, serviceItem := range s.ServiceSlice {
		if serviceItem.Info.LoadType != public.LoadTypeHTTP {
			continue
		}
		if serviceItem.HTTPRule.RuleType == public.HTTPRuleTypeDomain { //域名的匹配方式
			if serviceItem.HTTPRule.Rule == host {
				return serviceItem, nil
			}
		}
		if serviceItem.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL { //前缀的匹配方式
			if strings.HasPrefix(path, serviceItem.HTTPRule.Rule) {
				return serviceItem, nil
			}
		}
	}
	return nil, errors.New("not matched service")
}

// ServiceManager 的方法，用于一次性加载服务信息。
func (s *ServiceManager) LoadOnce() error {

	// 使用 sync.Once 确保该函数只被执行一次。
	s.init.Do(func() {

		// 创建一个新的 ServiceInfo 结构体实例。
		serviceInfo := &ServiceInfo{}

		// 创建一个测试上下文和记录器
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		// 从默认连接池中获取一个 GORM 数据库连接。
		tx, err := lib.GetGormPool("default")
		if err != nil {
			s.err = err // 如果有错误，保存错误到 ServiceManager 的 err 字段。d
			return
		}

		// 设置服务列表查询的参数。
		params := &dto.ServiceListInput{PageNo: 1, PageSize: 99999}

		// 获取服务列表。
		list, _, err := serviceInfo.PageList(c, tx, params)
		if err != nil {
			s.err = err // 如果有错误，保存错误到 ServiceManager 的 err 字段。
			return
		}

		// 加锁，保证线程安全。
		s.Locker.Lock()
		defer s.Locker.Unlock() // 使用 defer 确保在函数退出时解锁。

		// 遍历服务列表。
		for _, listItem := range list {
			tmpItem := listItem //

			// 获取每个服务的详细信息。
			serviceDetail, err := tmpItem.ServiceDetail(c, tx, &tmpItem)
			if err != nil {
				s.err = err // 如果有错误，保存错误到 ServiceManager 的 err 字段。
				return
			}

			// 保存服务的详细信息到 ServiceManager 的 ServiceMap 中，使用服务名称作为键。
			s.ServiceMap[listItem.ServiceName] = serviceDetail

			// 同时将服务的详细信息添加到 ServiceSlice 切片中。
			s.ServiceSlice = append(s.ServiceSlice, serviceDetail)
		}
	})

	// 返ServiceManager 的 err 字段，如果有错误的话。
	return s.err
}
