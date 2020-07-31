package umonitor

import (
	"go.uber.org/zap"
)

var (
	factories = make(map[string]func(*UAuth, *uZoneInfo, *ucloudResourceMetrics) (*ucloudResourceMetrics, error, string))
	selfConf  = umonConfNew()
)

// 采集基本配置
type umonConf struct {
	uauth  *UAuth
	uzone  *uZoneInfo
	logger *zap.Logger
	utypes *[]string
}

func umonConfNew() *umonConf {
	loglevel := "info"
	logpath := "./ucloud_exporter.log"

	logger := InitLogger(logpath, loglevel)
	typelist := make([]string, 0, len(factories))
	for key, _ := range factories {
		typelist = append(typelist, key)
	}

	return &umonConf{
		uauth:  nil,
		uzone:  nil,
		logger: logger,
		utypes: &typelist,
	}
}

// 配置初始化
func InitConf(pri, pub string, list *[]string, logger *zap.Logger) {
	defer logger.Sync()
	selfConf.logger = logger
	selfConf.initAuth(pri, pub)

	if nil != list {

		typelist := make([]string, 0, len(*selfConf.utypes))
		for _, key := range *list {
			_, found := factories[key]
			if found {
				typelist = append(typelist, key)
			}
		}

		selfConf.utypes = &typelist
	}

	logger.Debug(
		"collector type is ",
		zap.Any("typelist", selfConf.utypes),
	)
}

// ucloud认证初始化
func (umon *umonConf) initAuth(pri, pub string) {
	umon.uauth = uAuthNew(pri, pub)

	uzone, err := umon.uauth.GetBaseAccountZoneList()
	if nil != err {
		umon.logger.Panic(
			"get zone false.",
			zap.Any("msg", err),
		)
	}

	umon.uzone = uzone
}

// 注册函数
func registerResource(factoryName string, factory func(*UAuth, *uZoneInfo, *ucloudResourceMetrics) (*ucloudResourceMetrics, error, string)) {
	factories[factoryName] = factory
	selfConf.logger.Info("register func : ",
		zap.String("factoryName", string(factoryName)),
	)
}
