package main

import (
	"fmt"
	"net/http"
	"time"

	"sync"

	"github.com/gin-gonic/gin"
	// "golib/lang/pack"
	// "golib/lang/pack"
)

type PageLoadData struct {
	Meta struct {
		SendEventID      string `json:"sendEventID"`
		PageLocation     string `json:"pageLocation"`
		Host             string `json:"host"`
		Path             string `json:"path"`
		Query            string `json:"query"`
		Protocol         string `json:"protocol"`
		PageTitle        string `json:"pageTitle"`
		PCode            int    `json:"pCode"`
		ProjectAccessKey string `json:"projectAccessKey"`
		Screen           struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"screen"`
		SessionID string `json:"sessionID"`
		UserAgent string `json:"userAgent"`
		UserID    string `json:"userID"`
	} `json:"meta"`
	PageLoad struct {
		NavigationTiming struct {
			StartTimeStamp int64 `json:"startTimeStamp"`
			Data           struct {
				Redirect struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"redirect"`
				Cache struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"cache"`
				Connect struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"connect"`
				DNS struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"dns"`
				Ssl struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"ssl"`
				Download struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"download"`
				FirstByte struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"firstByte"`
				DomInteractive   int `json:"domInteractive"`
				DomContentLoaded struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"domContentLoaded"`
				DomComplete int `json:"domComplete"`
				DomLoad     struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"domLoad"`
				LoadTime     int `json:"loadTime"`
				BackendTime  int `json:"backendTime"`
				FrontendTime struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"frontendTime"`
				RenderTime struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"renderTime"`
			} `json:"data"`
			EventID string `json:"eventID"`
		} `json:"navigationTiming"`
		Resource []struct {
			StartTime      int    `json:"startTime"`
			StartTimeStamp int64  `json:"startTimeStamp"`
			EventID        string `json:"eventID"`
			Type           string `json:"type"`
			URL            string `json:"url"`
			URLHost        string `json:"urlHost"`
			URLPath        string `json:"urlPath"`
			URLQuery       string `json:"urlQuery"`
			URLProtocol    string `json:"urlProtocol"`
			Timing         struct {
				Redirect struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"redirect"`
				Cache struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"cache"`
				Connect struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"connect"`
				DNS struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"dns"`
				Ssl struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"ssl"`
				Download struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"download"`
				FirstByte struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"firstByte"`
				Duration int `json:"duration"`
				Size     int `json:"size"`
			} `json:"timing"`
			ResourceInfo struct {
				Method string `json:"method"`
				Status int    `json:"status"`
			} `json:"resourceInfo"`
			TraceInfo struct {
				MtID string `json:"mtID"`
				TxID string `json:"txID"`
			} `json:"traceInfo"`
		} `json:"resource"`
		TotalDuration int `json:"totalDuration"`
	} `json:"pageLoad"`
}

type PageLoadMap struct {
	sync.RWMutex
	items map[int64][]PageLoadData
	Stop  chan bool
}

func NewPageLoadMap() *PageLoadMap {
	urlMap := new(PageLoadMap)
	urlMap.Stop = make(chan bool)
	urlMap.SendPageLoadMapTagCountGoFunc()
	return urlMap
}

func (um *PageLoadMap) SendPageLoadMapTagCountGoFunc() {
	fiveSecondsTicker := time.NewTicker(10 * time.Second)
	now := time.Now().UTC()

	// 9초, 4초
	then := time.Date(now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second()+5-now.Second()%5+4, 0, time.UTC)
	// fmt.Println("then:", then)
	diff := then.Sub(now)
	// fmt.Println("diff:", diff)
	delay5s := time.NewTimer(diff)
	go func() {
		for {
			select {
			case <-delay5s.C:

				fiveSecondsTicker.Reset(5 * time.Second)
				delay5s.Stop()
				fmt.Println("delay5s.C:", time.Now())
				if um.items != nil {
					for k := range um.items {
						fmt.Println(k)
						fmt.Println(um.GetPcodeAvg(k))
						um.Remove(k)
					}
				}
			case <-fiveSecondsTicker.C:
				fmt.Println("5s", time.Now())
				if um.items != nil {
					for k := range um.items {
						fmt.Println(k)
						fmt.Println(um.GetPcodeAvg(k))
						um.Remove(k)
					}
				}

			case <-um.Stop:
				fmt.Println("stop: <-um.Stop", <-um.Stop)
				return
			}
		}
	}()
}

func (um *PageLoadMap) CloseUrlMap() {
	um.Lock()
	defer um.Unlock()
	if um.items != nil {
		for k := range um.items {
			delete(um.items, k)
		}
	}
	close(um.Stop)
}

func (um *PageLoadMap) Add(pCode int64, pgload PageLoadData) {
	um.Lock()
	defer um.Unlock()
	if um.items == nil {
		um.items = make(map[int64][]PageLoadData, 10)
	}
	um.items[pCode] = append(um.items[pCode], pgload)
}

func (um *PageLoadMap) Remove(pCode int64) {
	um.Lock()
	defer um.Unlock()
	if um.items != nil {
		delete(um.items, pCode)
	}
}

func (um *PageLoadMap) GetPcodeAvg(pCode int64) (int, int, int, int, int) {
	um.RLock()
	defer um.RUnlock()

	count := 0
	pgBackendSum := 0
	pgFrontendSum := 0
	pgDurationSum := 0
	pgLoadTimeSum := 0

	pgList, exists := um.items[pCode]
	if exists {
		for _, v := range pgList {
			pgBackendSum += v.PageLoad.NavigationTiming.Data.BackendTime
			pgFrontendSum += v.PageLoad.NavigationTiming.Data.FrontendTime.Duration
			pgDurationSum += v.PageLoad.TotalDuration
			pgLoadTimeSum += v.PageLoad.NavigationTiming.Data.LoadTime
		}
		count = len(pgList)
	}

	if count != 0 {
		return count, pgBackendSum / count, pgFrontendSum / count,
			pgDurationSum / count, pgLoadTimeSum / count
	}

	return count, pgBackendSum, pgFrontendSum, pgDurationSum, pgLoadTimeSum
}

func (um *PageLoadMap) GetDeviceAvg(pCode int64) (cnt int, avg int) {
	um.RLock()
	defer um.RUnlock()

	return 0, 0
}

var pgMap *PageLoadMap

func main() {

	r := gin.Default()

	pgMap = NewPageLoadMap()
	defer pgMap.CloseUrlMap()

	v1 := r.Group("/")
	{
		v1.POST("/pageLoad", pageLoad)
		v1.POST("/webVitals", webVitals)
		v1.POST("/resource", resource)
	}
	r.Run(":8848")
}

func pageLoad(c *gin.Context) {
	var pg PageLoadData // page performance timing data
	err := c.ShouldBindJSON(&pg)

	if err != nil {
		fmt.Println("pageLoad Error")
		c.JSON(http.StatusBadRequest, gin.H{"pageLoad Error": err.Error()})
		return
	}

	if pg.Meta.PCode == 0 {
		fmt.Println("pg.Meta.PCode == 0")
		c.JSON(http.StatusBadRequest, gin.H{"pCode == 0": "pg.Meta.PCode == 0"})
		return
	}

	pgMap.Add(int64(pg.Meta.PCode), pg)
	c.JSON(http.StatusOK, gin.H{
		"message": "pageLoad OK",
	})
}

func webVitals(c *gin.Context) {
	// time.Sleep(1000 * time.Microsecond)
	rawData, _ := c.GetRawData()
	strRawData := string(rawData)
	fmt.Println(strRawData)
	c.JSON(http.StatusOK, gin.H{
		"message": "POST webVitals OK",
	})
}

func resource(c *gin.Context) {
	rawData, _ := c.GetRawData()
	strRawData := string(rawData)
	fmt.Println(strRawData)
	// c.JSON(http.StatusOK, gin.H{
	// 	"message": "POST resource OK",
	// })
	c.String(http.StatusOK, strRawData)
}
