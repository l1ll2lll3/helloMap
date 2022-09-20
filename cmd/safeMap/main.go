package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"sync"
	"time"

	"github.com/whatap/golib/io"

	// "golib/lang/pack"
	"github.com/whatap/golib/lang/pack"
)

type UrlMap struct {
	sync.RWMutex
	items map[int64]map[string][]UrlRespTimes
	Stop  chan bool
}

func (um *UrlMap) SendUrlMapTagCountGoFunc() {
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
				if um.items != nil {
					sTime := time.Now()
					for k, _ := range um.items {
						cnt, pCodeAvg := um.GetPcodeAvg(k)
						p := pack.NewTagCountPack()
						p.Pcode = k
						p.Oid = 0
						// p.Time = pg.PageLoad.NavigationTiming.StartTimeStamp
						p.Time = time.Now().UnixMilli()
						p.Category = "rum_page_load_all_page"

						p.Tags.PutString("page_all", "-")
						p.Put("page_load_count", cnt)
						p.Put("page_load_duration", pCodeAvg.page_load_duration)
						p.Put("page_load_backend_time", pCodeAvg.page_load_backend_time)
						p.Put("page_load_frontend_time", pCodeAvg.page_load_frontend_time)
						err := Call("http://192.168.202.181:6633/merge/", p)
						if err != nil {
							fmt.Println(err)
						}
						uavgs := um.GetUrlAvgs(k)
						for _, v := range uavgs {
							// fmt.Println(k, v)
							p := pack.NewTagCountPack()
							p.Pcode = k
							p.Oid = 0
							// p.Time = pg.PageLoad.NavigationTiming.StartTimeStamp
							p.Time = time.Now().UnixMilli()
							p.Category = "rum_page_load_each_page"

							p.Tags.PutString("page_path", v.url)
							p.Put("page_load_count", v.count)
							p.Put("page_load_duration", v.avg.page_load_duration)
							p.Put("page_load_backend_time", v.avg.page_load_backend_time)
							p.Put("page_load_frontend_time", v.avg.page_load_frontend_time)
							err := Call("http://192.168.202.181:6633/merge/", p)
							if err != nil {
								fmt.Println(err)
							}
						}
						um.Remove(k)
					}
					eTime := time.Now()
					fmt.Println("duration for 5seonds http calls:", eTime.Sub(sTime))
				}

				// mapHitMapRumPack = make(map[int64]*pack.HitMapRumPack)
				// um.Unlock()
			case <-fiveSecondsTicker.C:
				fmt.Println(time.Now().UTC())
				if um.items != nil {
					sTime := time.Now()
					for k := range um.items {
						cnt, pCodeAvg := um.GetPcodeAvg(k)
						p := pack.NewTagCountPack()
						p.Pcode = k
						p.Oid = 0
						// p.Time = pg.PageLoad.NavigationTiming.StartTimeStamp
						p.Time = time.Now().UnixMilli()
						p.Category = "rum_page_load_all_page"

						p.Tags.PutString("page_all", "-")
						p.Put("page_load_count", cnt)
						p.Put("page_load_duration", pCodeAvg.page_load_duration)
						p.Put("page_load_backend_time", pCodeAvg.page_load_backend_time)
						p.Put("page_load_frontend_time", pCodeAvg.page_load_frontend_time)
						err := Call("http://192.168.202.181:6633/merge/", p)
						if err != nil {
							fmt.Println(err)
						}
						uavgs := um.GetUrlAvgs(k)
						for _, v := range uavgs {
							// fmt.Println(k, v)
							p := pack.NewTagCountPack()
							p.Pcode = k
							p.Oid = 0
							// p.Time = pg.PageLoad.NavigationTiming.StartTimeStamp
							p.Time = time.Now().UnixMilli()
							p.Category = "rum_page_load_each_page"

							p.Tags.PutString("page_path", v.url)
							p.Put("page_load_count", v.count)
							p.Put("page_load_duration", v.avg.page_load_duration)
							p.Put("page_load_backend_time", v.avg.page_load_backend_time)
							p.Put("page_load_frontend_time", v.avg.page_load_frontend_time)
							err := Call("http://192.168.202.181:6633/merge/", p)
							if err != nil {
								fmt.Println(err)
							}
						}
						um.Remove(k)
					}
					eTime := time.Now()
					fmt.Println("duration for 5seonds http calls:", eTime.Sub(sTime))
				} else {
					fmt.Println("没有")
				}
			case <-um.Stop:
				fmt.Println("stop: <-um.Stop", <-um.Stop)
				return
			}
		}
	}()

}

func (um *UrlMap) CloseUrlMap() {
	um.Lock()
	defer um.Unlock()
	if um.items != nil {
		for k, v := range um.items {
			for k1, _ := range v {
				v[k1] = nil
				delete(v, k1)
			}
			delete(um.items, k)
		}
	}
	close(um.Stop)
}

func NewUrlMap() *UrlMap {
	urlMap := new(UrlMap)
	urlMap.Stop = make(chan bool)
	urlMap.SendUrlMapTagCountGoFunc()
	return urlMap
}

type UrlAvg struct {
	url   string
	count int
	avg   UrlRespTimes
}

type UrlRespTimes struct {
	page_load_frontend_time float32
	page_load_backend_time  float32
	page_load_duration      float32
}

func (um *UrlMap) MakeUrlRespTimes(frontend, backend, duration int) (aaa UrlRespTimes) {
	aaa.page_load_frontend_time = float32(frontend)
	aaa.page_load_backend_time = float32(backend)
	aaa.page_load_duration = float32(duration)
	return aaa
}

func (um *UrlMap) Sum(urlRespTimes []UrlRespTimes) (total UrlRespTimes) {
	for _, v := range urlRespTimes {
		total.page_load_backend_time += v.page_load_backend_time
		total.page_load_duration += v.page_load_duration
		total.page_load_frontend_time += v.page_load_frontend_time
	}
	return total
}

func (um *UrlMap) Avg(urlRespTimes []UrlRespTimes) (urlRespAvg UrlRespTimes) {
	length := float32(len(urlRespTimes))
	urlRespAvg = um.Sum(urlRespTimes)
	urlRespAvg.page_load_backend_time = urlRespAvg.page_load_backend_time / length
	urlRespAvg.page_load_duration = urlRespAvg.page_load_duration / length
	urlRespAvg.page_load_frontend_time = urlRespAvg.page_load_frontend_time / length

	return urlRespAvg
}

func (um *UrlMap) Add(pCode int64, url string, duration UrlRespTimes) {
	um.Lock()
	defer um.Unlock()

	if um.items == nil {
		// fmt.Println("um.items == nil")
		um.items = make(map[int64]map[string][]UrlRespTimes, 10)
	}
	if um.items[pCode] == nil {
		// fmt.Println("um.items[pCode] == nil")
		um.items[pCode] = make(map[string][]UrlRespTimes, 10)
	}
	if um.items[pCode][url] == nil {
		// fmt.Println("um.items[pCode][url] == nil")
		um.items[pCode][url] = make([]UrlRespTimes, 0, 10)
	}
	um.items[pCode][url] = append(um.items[pCode][url], duration)
}

func (um *UrlMap) GetUrlAvg(pCode int64, urlkey string) (int, UrlRespTimes) {
	um.RLock()
	defer um.RUnlock()
	count := 0
	avg := UrlRespTimes{0, 0, 0}
	// sum := UrlRespTimes{0, 0, 0}
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

func (um *UrlMap) GetUrls(pCode int64) []string {
	um.RLock()
	defer um.RUnlock()
	strKeys := []string{}
	_, pexists := um.items[pCode]
	if pexists {
		for k, _ := range um.items[pCode] {
			strKeys = append(strKeys, k)
		}
	}
	return strKeys
}

func (um *UrlMap) GetUrlAvgs(pCode int64) (urlavgs []UrlAvg) {
	um.RLock()
	defer um.RUnlock()
	// urlavgs := []UrlAvg{}
	_, pexists := um.items[pCode]
	if pexists {
		for k, v := range um.items[pCode] {
			urlavgs = append(urlavgs, UrlAvg{k, len(v), um.Avg(v)})
		}
	}
	return urlavgs
}

func (um *UrlMap) GetPcodeAvg(pCode int64) (int, UrlRespTimes) {
	um.RLock()
	defer um.RUnlock()
	count := 0
	avg := UrlRespTimes{0, 0, 0}
	var sums []UrlRespTimes
	_, pexists := um.items[pCode]
	if pexists {
		for _, v := range um.items[pCode] {
			sums = append(sums, um.Sum(v))
			count += len(v)
		}
		if count != 0 {
			avg = um.Sum(sums)
			avg.page_load_backend_time = avg.page_load_backend_time / float32(count)
			avg.page_load_duration = avg.page_load_duration / float32(count)
			avg.page_load_frontend_time = avg.page_load_frontend_time / float32(count)
		}
	}
	return count, avg
}

func (um *UrlMap) Remove(pCode int64) {
	um.Lock()
	defer um.Unlock()
	if um.items != nil {
		v, pexists := um.items[pCode]
		if pexists {
			for k := range v {
				v[k] = nil
				delete(v, k)
			}
		}
		delete(um.items, pCode)
	}
}

func (um *UrlMap) GetMapDump() (strDump string) {
	um.RLock()
	defer um.RUnlock()
	strDump = fmt.Sprintf("%+v", um.items)
	return strDump
}

func Call(urlPath string, p pack.Pack) error {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	b := io.NewDataOutputX()
	p.Write(b)
	// fmt.Println(b)
	// b, err := ioutil.ReadFile("photo.png")
	// if err != nil {
	// 	return err
	// }

	req, err := http.NewRequest("POST", urlPath, bytes.NewReader(b.ToByteArray()))
	if err != nil {
		log.Println(err)
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	rsp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	if rsp.StatusCode != http.StatusOK {
		log.Printf("Request failed with response code: %d", rsp.StatusCode)
	}

	return nil
}

func main() {

	// urlMap := make(map[int64]map[string][]int, 100)
	// urlMap[111] = make(map[string][]int, 100)
	// urlMap[111]["url"] = make([]int, 10)

	um := NewUrlMap()
	um.Add(222, "yahoo.com", UrlRespTimes{12, 13, 25})
	um.Add(222, "kr.yahcvbnoo.com", UrlRespTimes{15, 16, 31})
	um.Add(222, "kr.ydfghjahoo.com", UrlRespTimes{16, 17, 33})
	um.Add(222, "yahoo.com", UrlRespTimes{16, 17, 33})
	// um.HitMapRumPackGoFunc()

	fmt.Println("Sum of 222 yahoo", um.Sum(um.items[222]["yahoo.com"]))

	for i := 0; i < 17; i++ {
		um.Add(222, fmt.Sprint("kr.ya.com", i), UrlRespTimes{2 * float32(i), 3 * float32(i), 5 * float32(i)})
	}

	fmt.Println("langth of um.items", len(um.items))
	fmt.Println("langth of 222", len(um.items[222]))
	fmt.Println("langth of [222][kr.ya.com0]", len(um.items[222]["kr.ya.com0"]))

	delete(um.items[222], "kr.ya.com")
	um.Add(222, "kr.ya.com", UrlRespTimes{19, 18, 38})
	fmt.Println("langth of [222][kr.ya.com]", len(um.items[222]["kr.ya.com"]))

	UrlAvgs := um.GetUrlAvgs(222)
	for i, v := range UrlAvgs {
		fmt.Println(i, v)
	}

	Urls := um.GetUrls(222)
	for i, v := range Urls {
		fmt.Println(i, "url:", v)
	}

	cnt, pcodeavg := um.GetPcodeAvg(222)
	fmt.Println("Pcode Avg:", pcodeavg, "cnt:", cnt)

	// um.Remove(222)

	// fmt.Println("remove 222:")
	// cnt, pcodeavg = um.GetPcodeAvg(222)
	// fmt.Println("Pcode Avg:", pcodeavg, "cnt:", cnt)

	p := pack.NewTagCountPack()
	p.Pcode = int64(41)
	p.Oid = 0
	// p.Time = pg.PageLoad.NavigationTiming.StartTimeStamp
	p.Time = time.Now().Unix()
	p.Category = "rum_page_load_device_each_page"
	// browser, _, _ := ParseUserAgent(pg.Meta.UserAgent)
	// uaparse := ua.Parse(pg.Meta.UserAgent)
	p.Tags.PutString("device_name", "Desktop")
	// fmt.Println("device_name", uaparse.Device)
	p.Tags.PutString("page_path", "http://kr.yao.com")

	// loadTime, backendTime
	p.Put("page_load_count", 1)
	p.Put("page_load_duration", 22)

	// oneerr := onewayClient.Send(p, wnet.WithLicense(pg.Meta.ProjectAccessKey))
	// if oneerr != nil {
	// 	fmt.Println(oneerr)
	// }

	sTime := time.Now()
	// dout := io.NewDataOutputX()
	// p.Write(dout)
	// body := bytes.NewBuffer(dout.ToByteArray())
	// http.Post("http://192.168.202.181:6633/merge", "application/octet-stream", body)
	err := Call("http://192.168.202.181:6633/merge/", p)
	if err != nil {
		fmt.Println(err)
	}
	eTime := time.Now()
	fmt.Println("duration for http call:", eTime.Sub(sTime))

	time.Sleep(time.Second * 10)
	fmt.Println("sleep 10s end")

	fmt.Println((time.Now().UnixNano() / 1000000), time.Now().UnixMilli())

	um.CloseUrlMap()

	time.Sleep(time.Second * 2)
	fmt.Println("sleep 5s end")

	lll := um.MakeUrlRespTimes(111, 222, 333)
	fmt.Printf("%f,%f,%f", lll.page_load_backend_time, lll.page_load_duration, lll.page_load_frontend_time)

}
