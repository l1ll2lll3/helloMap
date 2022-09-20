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

type PlusUrlMap struct {
	Type string // device, os, browser
	sync.RWMutex
	items map[string]map[int64]map[string][]float32
	Stop  chan bool
}

func (um *PlusUrlMap) SendTagCounter() {
	if um.items != nil {
		for k, v := range um.items {
			for k1 := range v {
				cnt, pCodeAvg := um.GetPcodeAvg(k, k1)
				p := pack.NewTagCountPack()
				p.Pcode = k1
				p.Oid = 0
				p.Time = time.Now().UnixMilli()
				p.Category = "rum_page_load_device_all_page"

				p.Tags.PutString("device_name", k)
				p.Put("page_load_count", cnt)
				p.Put("page_load_duration", pCodeAvg)
				err := Call("http://192.168.202.181:6633/merge/", p)
				if err != nil {
					fmt.Println(err)
				}

				uavgs := um.GetUrlAvgs(k, k1)
				for _, v1 := range uavgs {
					// fmt.Println(k, v)
					p := pack.NewTagCountPack()
					p.Pcode = k1
					p.Oid = 0
					// p.Time = pg.PageLoad.NavigationTiming.StartTimeStamp
					p.Time = time.Now().UnixMilli()
					p.Category = "rum_page_load_device_each_page"

					p.Tags.PutString("device_name", k)
					p.Tags.PutString("page_path", v1.url)
					p.Put("page_load_count", cnt)
					p.Put("page_load_duration", pCodeAvg)
					err := Call("http://192.168.202.181:6633/merge/", p)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}
		um.RemoveAll()
	}
}

func (um *PlusUrlMap) SendPlusUrlMapTagCountGoFunc() {
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
				um.SendTagCounter()
			case <-fiveSecondsTicker.C:
				fmt.Println(time.Now().UTC())
				um.SendTagCounter()
			case <-um.Stop:
				fmt.Println("stop: <-um.Stop", <-um.Stop)
				return
			}
		}
	}()

}

func (um *PlusUrlMap) ClosePlusUrlMap() {
	um.RemoveAll()
	close(um.Stop)
}

func NewPlusUrlMap(Type string) *PlusUrlMap {
	plusUrlMap := new(PlusUrlMap)
	plusUrlMap.Type = Type
	plusUrlMap.Stop = make(chan bool)
	plusUrlMap.SendPlusUrlMapTagCountGoFunc()
	return plusUrlMap
}

type PlusUrlAvg struct {
	url   string
	count int
	avg   float32
}

func (um *PlusUrlMap) Sum(durations []float32) (total float32) {
	total = 0
	for _, v := range durations {
		total += v
	}
	return total
}

func (um *PlusUrlMap) Avg(durations []float32) (avg float32) {
	length := len(durations)
	avg = -1 // error
	if length != 0 {
		avg = um.Sum(durations) / float32(length)
	}
	return avg
}

func (um *PlusUrlMap) Add(plus string, pCode int64, url string, duration float32) {
	um.Lock()
	defer um.Unlock()

	if um.items == nil {
		// fmt.Println("um.items == nil")
		um.items = make(map[string]map[int64]map[string][]float32, 10)
	}
	if um.items[plus] == nil {
		// fmt.Println("um.items[pCode] == nil")
		um.items[plus] = make(map[int64]map[string][]float32, 10)
	}
	if um.items[plus][pCode] == nil {
		// fmt.Println("um.items[pCode][url] == nil")
		um.items[plus][pCode] = make(map[string][]float32, 10)
	}
	if um.items[plus][pCode][url] == nil {
		// fmt.Println("um.items[pCode][url] == nil")
		um.items[plus][pCode][url] = make([]float32, 0, 10)
	}
	um.items[plus][pCode][url] = append(um.items[plus][pCode][url], duration)
}

func (um *PlusUrlMap) GetUrlAvgs(plus string, pCode int64) (plusurlavgs []PlusUrlAvg) {
	um.RLock()
	defer um.RUnlock()
	// urlavgs := []UrlAvg{}
	if um.items != nil {
		for k, v := range um.items[plus][pCode] {
			plusurlavgs = append(plusurlavgs, PlusUrlAvg{k, len(v), um.Avg(v)})
		}
	}

	return plusurlavgs
}

func (um *PlusUrlMap) GetPcodeAvg(plus string, pCode int64) (cnt int, avg float32) {
	um.RLock()
	defer um.RUnlock()
	cnt = 0  // error
	avg = -1 // error
	sum := float32(0)

	if um.items != nil {
		for _, v := range um.items[plus][pCode] {
			sum += um.Sum(v)
			cnt += len(v)
		}
	}
	if cnt != 0 {
		avg = sum / float32(cnt)
	}
	return cnt, avg
}

func (um *PlusUrlMap) RemoveAll() {
	um.Lock()
	defer um.Unlock()
	if um.items != nil {
		for k, v := range um.items {
			for k1, v1 := range v {
				for k2, _ := range v1 {
					v1[k2] = nil
					delete(v1, k2)
				}
				delete(v, k1)
			}
			delete(um.items, k)
		}
	}
}

func (um *PlusUrlMap) Remove(plus string, pCode int64) {
	um.Lock()
	defer um.Unlock()
	if um.items != nil {
		for k := range um.items[plus][pCode] {
			um.items[plus][pCode][k] = nil
			delete(um.items[plus][pCode], k)
		}
		delete(um.items[plus], pCode)
	}
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
	sTime := time.Now()
	dum := NewPlusUrlMap("device")

	dum.Add("desktop", 1, "yahoo.com", 33)
	dum.Add("desktop", 226, "yahoo.com", 34)
	dum.Add("tablelet", 232, "yahoo.com", 35)
	dum.Add("desktop", 226, "yahoo.com", 39.3)
	// um.HitMapRumPackGoFunc()

	for j := 22; j < 10203; j++ {
		for i := 1; i < 17; i++ {
			dum.Add("desktop", 222, fmt.Sprint("kr.ya.com", i), (12.8+float32(i)*float32(i))/float32(j))
		}
	}

	eTime := time.Now()
	fmt.Println("duration for insert:", eTime.Sub(sTime))

	fmt.Println("langth of um.items", len(dum.items))
	fmt.Println("langth of desktop 222", len(dum.items["desktop"][222]))

	// fmt.Println("dum:", dum)
	// fmt.Println("dum.items:", dum.items)
	// fmt.Println("dum.items[desktop]:", dum.items["desktop"])
	// fmt.Println("dum.items[desktop][222]:", dum.items["desktop"][222])

	// fmt.Printf("langth of [222][kr.ya.com0]", len(um.items[222]["kr.ya.com0"]))

	// delete(um.items[222], "kr.ya.com")
	// um.Add(222, "kr.ya.com", UrlRespTimes{19, 18, 38})
	// fmt.Println("langth of [222][kr.ya.com]", len(um.items[222]["kr.ya.com"]))

	UrlAvgs := dum.GetUrlAvgs("desktop", 222)
	for i, v := range UrlAvgs {
		fmt.Println("GetUrlAvgs", i, v)
	}

	// Urls := um.GetUrls(222)
	// for i, v := range Urls {
	// 	fmt.Println(i, "url:", v)
	// }

	cnt, pcodeavg := dum.GetPcodeAvg("desktop", 222)
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
	// dout := io.NewDataOutputX()
	// p.Write(dout)
	// body := bytes.NewBuffer(dout.ToByteArray())
	// http.Post("http://192.168.202.181:6633/merge", "application/octet-stream", body)

	err := Call("http://192.168.202.181:6633/merge/", p)
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(time.Second * 10)
	fmt.Println("sleep 10s end")

	fmt.Println((time.Now().UnixNano() / 1000000), time.Now().UnixMilli())

	dum.ClosePlusUrlMap()

	time.Sleep(time.Second * 2)
	fmt.Println("sleep 5s end")
}
