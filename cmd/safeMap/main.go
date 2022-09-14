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

func (um *UrlMap) HitMapRumPackGoFunc() {
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
				fmt.Println("delay5s.C: send HitmapPack,", time.Now().UTC())
				// fmt.Printf("delay5s %+v\n", mapHitMapRumPack)
				um.Lock()
				// mapHitMapRumPack = make(map[int64]*pack.HitMapRumPack)
				um.Unlock()
			case <-fiveSecondsTicker.C:
				fmt.Println("fiveSecondsTicker.C: send HitmapPack,", time.Now().UTC())
				um.Lock()
				// for k := range mapHitMapRumPack {
				// 	onewayClient.Send(mapHitMapRumPack[k], wnet.WithLicense(mapHitMapRumPack[k].AccessKey)) // Hitmap
				// 	// fmt.Printf("fiveSecondsTicker.C:%d %+v\n", k, mapHitMapRumPack[k])
				// 	// delete(mapHitMapRumPack, k)
				// }
				// mapHitMapRumPack = make(map[int64]*pack.HitMapRumPack)
				um.Unlock()
			case <-um.Stop:
				fmt.Println("stop: <-um.Stop", <-um.Stop)
				return
			}
		}
	}()

}

func NewUrlMap() *UrlMap {
	urlMap := new(UrlMap)
	urlMap.Stop = make(chan bool)
	urlMap.HitMapRumPackGoFunc()
	return urlMap
}

type UrlAvg struct {
	url   string
	count int
	avgs  UrlRespTimes
}

type UrlRespTimes struct {
	page_load_frontend_time float32
	page_load_backend_time  float32
	page_load_duration      float32
}

func (um *UrlMap) Sum(urlRespTimes []UrlRespTimes) (total UrlRespTimes) {
	for i := range urlRespTimes {
		total.page_load_backend_time += urlRespTimes[i].page_load_backend_time
		total.page_load_duration += urlRespTimes[i].page_load_duration
		total.page_load_frontend_time += urlRespTimes[i].page_load_frontend_time
	}
	return total
}

func (um *UrlMap) Avg(urlRespTimes []UrlRespTimes) (urlRespAvgs UrlRespTimes) {
	length := len(urlRespTimes)
	urlRespAvgs = um.Sum(urlRespTimes)
	urlRespAvgs.page_load_backend_time = urlRespAvgs.page_load_backend_time / float32(length)
	urlRespAvgs.page_load_duration = urlRespAvgs.page_load_duration / float32(length)
	urlRespAvgs.page_load_frontend_time = urlRespAvgs.page_load_frontend_time / float32(length)

	return urlRespAvgs
}

func (um *UrlMap) Add(pCode int64, url string, duration UrlRespTimes) {
	um.Lock()
	defer um.Unlock()

	if um.items == nil {
		fmt.Println("um.items == nil")
		um.items = make(map[int64]map[string][]UrlRespTimes, 10)
	}
	if um.items[pCode] == nil {
		fmt.Println("um.items[pCode] == nil")
		um.items[pCode] = make(map[string][]UrlRespTimes, 10)
	}
	if um.items[pCode][url] == nil {
		fmt.Println("um.items[pCode][url] == nil")
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

func (um *UrlMap) GetUrlKeyAvgs(pCode int64) (urlavgs []UrlAvg) {
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
		if len(sums) != 0 {
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
	delete(um.items, pCode)
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
	um.Add(111, "yahoo.com", UrlRespTimes{12, 13, 25})
	um.Add(333, "kr.yahoo.com", UrlRespTimes{15, 16, 31})
	um.Add(333, "kr.yahoo.com", UrlRespTimes{16, 17, 33})

	// um.HitMapRumPackGoFunc()

	for i := 0; i < 17; i++ {
		um.Add(222, fmt.Sprint("kr.ya.com", i), UrlRespTimes{2 * float32(i), 3 * float32(i), 5 * float32(i)})
	}

	fmt.Printf("%+v\n", um.items)

	fmt.Println("langth of um.items", len(um.items))
	fmt.Println("langth of 222", len(um.items[222]))
	fmt.Println("langth of [222][kr.ya.com0]", len(um.items[222]["kr.ya.com0"]))

	delete(um.items[222], "kr.ya.com")
	um.Add(222, "kr.ya.com", UrlRespTimes{19, 18, 38})
	// fmt.Printf("%+v\n", um.items)
	fmt.Println("langth of [222][kr.ya.com]", len(um.items[222]["kr.ya.com"]))

	// for _, v := range um.GetStrKeys(222) {
	// 	count, avg := um.GetUrlAvg(222, v)
	// 	fmt.Println("um.GetUrlAvg(222, kr.ya.com) count:", count, "avg:", avg, "url:", v)
	// }

	fmt.Println("Hello World")
	UrlAvgs := um.GetUrlKeyAvgs(222)
	for _, v := range UrlAvgs {
		fmt.Println(v)
	}

	_, pcodeavg := um.GetPcodeAvg(222)
	fmt.Println(pcodeavg)

	fmt.Println("Hello World")
	fmt.Println(um.GetMapDump())

	fmt.Println("Hello World 888")
	UrlAvgs = um.GetUrlKeyAvgs(888)
	if len(UrlAvgs) == 0 {
		fmt.Println("Hello World 888 len 0")
	}
	for _, v := range UrlAvgs {
		fmt.Println(v)
	}
	fmt.Println("Hello World 888 end")
	fmt.Println(um.GetMapDump())

	for i := 0; i < 17; i++ {
		um.Add(888, fmt.Sprint("kr.ya.com", i), UrlRespTimes{2 * float32(i), 3 * float32(i), 5 * float32(i)})
	}

	count, avg := um.GetPcodeAvg(888)
	fmt.Println("um.GetPcodeAvg(888) count:", count, "avg:", avg)

	count, avg = um.GetPcodeAvg(222)
	fmt.Println("um.GetPcodeAvg(222) count:", count, "avg:", avg)

	um.Remove(222)
	count, avg = um.GetPcodeAvg(222)
	fmt.Println("um.GetPcodeAvg(222) count:", count, "avg:", avg)

	um.Add(222, "kr.yb.com", UrlRespTimes{20, 20, 40})
	um.Add(222, "kr.ya.com", UrlRespTimes{22, 22, 44})
	count, avg = um.GetPcodeAvg(222)
	fmt.Printf("%+v\num.GetPcodeAvg(222) count:%d avg:%f\n", um.items, count, avg)

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

	close(um.Stop)
	fmt.Println("close end")

	time.Sleep(time.Second * 2)
	fmt.Println("sleep 5s end")

}
