package main

import (
	"fmt"
	"reflect"
	"time"
)

type lll struct {
	Elapsed []int
}

func main() {
	// oneHourTicker := time.NewTicker(2 * time.Hour)
	// now := time.Now().UTC()
	// then := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, time.UTC)
	// fmt.Println("then:", then)
	// diff := then.Sub(now)
	// fmt.Println("diff:", diff)
	// delay := time.NewTimer(diff)
	// fmt.Println("delay:", delay)
	// stop := make(chan bool)
	// go func() {
	// 	// c := 0
	// 	for {
	// 		select {
	// 		case <-delay.C:
	// 			delay.Stop()
	// 			fmt.Println("time.After(diff):", then.Hour(), "o'clock.")
	// 			fmt.Println("Reset")
	// 			fmt.Println("make timeticker")
	// 			oneHourTicker.Reset(time.Hour)
	// 			// timeticker = time.NewTicker()
	// 		case <-oneHourTicker.C:
	// 			fmt.Println("timeticker.C:", oneHourTicker)
	// 			fmt.Println("Reset")
	// 		case <-stop:
	// 			fmt.Println("stop")
	// 			return
	// 		}
	// 	}
	// }()

	qqq := true
	fmt.Println("sizeof qqq:", reflect.TypeOf(qqq).Size())
	world := make(map[int64]lll)
	var hello lll
	hello.Elapsed = append(hello.Elapsed, 1, 2, 3, 6)
	world[int64(12)] = hello
	fmt.Println(world)

	now := time.Now().UTC()
	fmt.Println("now", now)
	then := time.Date(now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second()+5-now.Second()%5, 0, time.UTC)

	fmt.Println("then", then)
	diff := then.Sub(now)
	fmt.Println("diff:", diff.Milliseconds())

	fiveSecondsTicker := time.NewTicker(10 * time.Second)
	stop := make(chan bool)
	now = time.Now().UTC()
	then = time.Date(now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second()+5-now.Second()%5, 0, time.UTC)
	fmt.Println("then:", then)
	diff = then.Sub(now)
	fmt.Println("diff:", diff)
	delay5s := time.NewTimer(diff)
	go func() {
		for {
			select {
			case <-delay5s.C:
				fiveSecondsTicker.Reset(5 * time.Second)
				fmt.Println("2 delay5s.C: send HitmapPack,", time.Now().UTC())
				delay5s.Stop()
			case <-fiveSecondsTicker.C:
				fmt.Println("2 fiveSecondsTicker.C: send HitmapPack,", time.Now().UTC())
			case <-stop:
				fmt.Println("stop:", <-stop)
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-delay5s.C:
				fiveSecondsTicker.Reset(5 * time.Second)
				fmt.Println("1 delay5s.C: send HitmapPack,", time.Now().UTC())
				delay5s.Stop()
			case <-fiveSecondsTicker.C:
				fmt.Println("1 fiveSecondsTicker.C: send HitmapPack,", time.Now().UTC())
			case <-stop:
				fmt.Println("stop:", <-stop)
				return
			}
		}
	}()

	for {
		time.Sleep(30 * time.Second)
		stop <- false
		stop <- true
	}
}
