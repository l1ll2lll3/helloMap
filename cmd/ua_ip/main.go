package main

import (
	// "github.com/ip2location/ip2location-go" //ip2location $199/year
	"fmt"

	"github.com/ip2location/ip2location-go/v9" //ip2location $199/year
	ua "github.com/mileusna/useragent"
)

func main() {
	// ua: {Chrome 103.0.5060.134 Windows 10.0  false false true false  Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.134 Safari/537.36}

	var userAgents = []string{
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36 OPR/38.0.2220.41",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.134 Safari/537.36",
		"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36 OPR/38.0.2220.41"}

	for i := range userAgents {
		// userAgent :=
		browser, os, device := ParseUserAgent(userAgents[i])
		fmt.Println("browser:", browser, "OS:", os, "Device:", device)
	}

	// if ua.Desktop {
	// 	fmt.Println("Device", "Desktop")
	// } else {
	// 	if ua.Mobile {
	// 		fmt.Println("Device", "Mobile")
	// 	} else {
	// 		fmt.Println("Device", "Tablet")
	// 	}
	// }

	db, err := ip2location.OpenDB("./IP2LOCATION-LITE-DB3.BIN")
	if err != nil {
		fmt.Print(err)
	}

	var ips = []string{"106.10.248.150", "13.124.123.77", "223.130.200.104"}

	for i := range ips {
		ip := ips[i]
		results, _ := db.Get_all(ip)
		fmt.Printf("%+v\n", results)
		fmt.Println("City:", results.City)
		fmt.Println("Region:", results.Region)
		fmt.Println("Country:", results.Country_short)
	}
	db.Close()
}

func ParseUserAgent(userAgent string) (string, string, string) {
	uaParse := ua.Parse(userAgent)
	var Device string
	switch {
	case uaParse.Desktop:
		Device = "Desktop"
	case uaParse.Bot:
		Device = "Bot"
	case uaParse.Mobile:
		Device = "Mobile"
	case uaParse.Tablet:
		Device = "Tablet"
	default:
		Device = "Unknown"
	}
	return uaParse.Name, uaParse.OS, Device
}
