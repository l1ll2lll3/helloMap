package main

import (
	"fmt"
	"sync"
	"time"
	// "golib/lang/pack"
)

type WebVitalsData struct {
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
	WebVitals struct {
		Cls float64 `json:"CLS"`
		Fid int     `json:"FID"`
		Lcp int     `json:"LCP"`
	} `json:"webVitals"`
}

type WebVitalsMap struct {
	sync.RWMutex
	items map[int64]map[string][]WebVitalsRespTimes
	Stop  chan bool
}

func (um *WebVitalsMap) SendWebVitalsTagCountGoFunc() {
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
				fmt.Println("delay5s.C:", time.Now().UTC())
				// um.Lock()

				// mapHitMapRumPack = make(map[int64]*pack.HitMapRumPack)
				// um.Unlock()
			case <-fiveSecondsTicker.C:
				fmt.Println(time.Now().UTC())

			case <-um.Stop:
				fmt.Println("stop: <-um.Stop", <-um.Stop)
				return
			}
		}
	}()

}

func (um *WebVitalsMap) CloseUrlMap() {
	um.Lock()
	defer um.Unlock()
	if um.items != nil {
		for k, v := range um.items {
			for k1 := range v {
				v[k1] = nil
				delete(v, k1)
			}
			delete(um.items, k)
		}
	}
	close(um.Stop)
}

func NewWebVitalsMap() *WebVitalsMap {
	webVitalsMap := new(WebVitalsMap)
	webVitalsMap.Stop = make(chan bool)
	webVitalsMap.SendWebVitalsTagCountGoFunc()
	return webVitalsMap
}

type WebVitalsMapAvg struct {
	url   string
	count int
	avg   WebVitalsRespTimes
}

type WebVitalsRespTimes struct {
	cls float32
	fid float32
	lcp float32
}

func (um *WebVitalsMap) MakeWebVitalsRespTimes(cls float64, fid, lcp int) (aaa WebVitalsRespTimes) {
	aaa.cls = float32(cls)
	aaa.fid = float32(fid)
	aaa.lcp = float32(lcp)
	return aaa
}

func (um *WebVitalsMap) Sum(WebVitalsRespTimes []WebVitalsRespTimes) (total WebVitalsRespTimes) {
	for _, v := range WebVitalsRespTimes {
		total.fid += v.fid
		total.lcp += v.lcp
		total.cls += v.cls
	}
	return total
}

func (um *WebVitalsMap) Avg(WebVitalsRespTimes []WebVitalsRespTimes) (urlRespAvg WebVitalsRespTimes) {
	length := len(WebVitalsRespTimes)
	if length != 0 {
		urlRespAvg = um.Sum(WebVitalsRespTimes)
		urlRespAvg.fid = urlRespAvg.fid / float32(length)
		urlRespAvg.lcp = urlRespAvg.lcp / float32(length)
		urlRespAvg.cls = urlRespAvg.cls / float32(length)
	}
	return urlRespAvg
}

func (um *WebVitalsMap) Add(pCode int64, url string, wVRespTimes WebVitalsRespTimes) {
	um.Lock()
	defer um.Unlock()

	if um.items == nil {
		// fmt.Println("um.items == nil")
		um.items = make(map[int64]map[string][]WebVitalsRespTimes, 10)
	}
	if um.items[pCode] == nil {
		// fmt.Println("um.items[pCode] == nil")
		um.items[pCode] = make(map[string][]WebVitalsRespTimes, 10)
	}
	if um.items[pCode][url] == nil {
		// fmt.Println("um.items[pCode][url] == nil")
		um.items[pCode][url] = make([]WebVitalsRespTimes, 0, 10)
	}
	um.items[pCode][url] = append(um.items[pCode][url], wVRespTimes)
}

func (um *WebVitalsMap) GetUrlAvg(pCode int64, urlkey string) (int, WebVitalsRespTimes) {
	um.RLock()
	defer um.RUnlock()
	count := 0
	avg := WebVitalsRespTimes{0, 0, 0}
	// sum := WebVitalsRespTimes{0, 0, 0}
	_, pexists := um.items[pCode]
	if pexists {
		val, uexists := um.items[pCode][urlkey]
		if uexists {
			count = len(val)
			if count != 0 {
				// sum = um.Sum(val)
				avg = um.Avg(val)
			}
		}
	}
	return count, avg
}

func (um *WebVitalsMap) GetUrls(pCode int64) []string {
	um.RLock()
	defer um.RUnlock()
	strKeys := []string{}
	_, pexists := um.items[pCode]
	if pexists {
		for k := range um.items[pCode] {
			strKeys = append(strKeys, k)
		}
	}
	return strKeys
}

func (um *WebVitalsMap) GetUrlAvgs(pCode int64) (webVitalsMapAvg []WebVitalsMapAvg) {
	um.RLock()
	defer um.RUnlock()
	// urlavgs := []UrlAvg{}
	_, pexists := um.items[pCode]
	if pexists {
		for k, v := range um.items[pCode] {
			webVitalsMapAvg = append(webVitalsMapAvg, WebVitalsMapAvg{k, len(v), um.Avg(v)})
		}
	}
	return webVitalsMapAvg
}

func (um *WebVitalsMap) GetPcodeAvg(pCode int64) (int, WebVitalsRespTimes) {
	um.RLock()
	defer um.RUnlock()
	count := 0
	avg := WebVitalsRespTimes{0, 0, 0}
	var sums []WebVitalsRespTimes
	_, pexists := um.items[pCode]
	if pexists {
		for _, v := range um.items[pCode] {
			sums = append(sums, um.Sum(v))
			count += len(v)
		}
		if count != 0 {
			avg = um.Sum(sums)
			avg.fid = avg.fid / float32(count)
			avg.lcp = avg.lcp / float32(count)
			avg.cls = avg.cls / float32(count)
		}
	}
	return count, avg
}

func (um *WebVitalsMap) Remove(pCode int64) {
	um.Lock()
	defer um.Unlock()
	if um.items != nil {
		v, pexists := um.items[pCode]
		if pexists {
			for k, _ := range v {
				v[k] = nil
				delete(v, k)
			}
		}
		delete(um.items, pCode)
	}
}

func (um *WebVitalsMap) GetMapDump() (strDump string) {
	um.RLock()
	defer um.RUnlock()
	strDump = fmt.Sprintf("%+v", um.items)
	return strDump
}
