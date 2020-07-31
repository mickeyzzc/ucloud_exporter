package umonitor

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	selfResource = ucloudResourcesNew()
)

type ucloudResources struct {
	resourceList map[string]*ucloudResourceMetrics
	sync.RWMutex
}

func ucloudResourcesNew() *ucloudResources {
	return &ucloudResources{
		resourceList: make(map[string]*ucloudResourceMetrics),
	}
}

type ucloudResourceMetrics struct {
	// eg : uhost
	ResourceType   *ucloudMetrics
	ResourceIDList map[string]*resourceLabels
	TimeRange      int
	sync.RWMutex
}

type resourceLabels struct {
	hashid        string
	project_id    string
	project_name  string
	region_id     string
	zone_id       string
	resource_name string
	resource_id   string
	resource_type string
}

func ResourceHandle(renetTime *int64) {
	resourceUpdate()
	gcTimeChan := make(chan *int64)
	//go timeTick(tick)
	go func(ts *int64) {
		for {
			time.Sleep(time.Duration(*ts) * time.Second)
			t := time.Now().Unix() - *ts
			gcTimeChan <- &t
		}
	}(renetTime)
	for {
		select {
		case <-gcTimeChan:
			resourceUpdate()
		}
	}
}

func resourceUpdate() {
	selfResource.Lock()
	defer selfResource.Unlock()
	wg := sync.WaitGroup{}
	wg.Add(len(factories))
	for name, fn := range factories {
		go func(selfname string, selffn func(*UAuth, *uZoneInfo, *ucloudResourceMetrics) (*ucloudResourceMetrics, error, string)) {
			defer wg.Done()
			resources, err, fname := selffn(selfConf.uauth, selfConf.uzone, nil)
			selfConf.logger.Debug(
				"resource update now",
				zap.String("type", selfname),
				zap.String("fn", fname),
			)

			if err != nil {
				return
			}
			selfResource.resourceList[selfname] = resources
		}(name, fn)
	}

	wg.Wait()
	selfConf.logger.Info(
		"resource update ok",
	)
}
