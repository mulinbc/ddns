package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

const confFilePath = "conf/conf.json"

func main() {

	ddns, err := configInit()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(ddnsConf)

	err = ddns.init()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(ddnsConf.DNSRecord)

	ipChan := make(chan string)
	go ddns.GetIP.getLocalIP(ipChan)

	oldIP := ""
	newIP := ""
	for {
		newIP = <-ipChan
		if newIP != oldIP {
			ddns.updateDNSRecord(newIP)
			oldIP = newIP
		}
		// fmt.Println("old =new")
	}
}

func configInit() (ddnsConf ddns, err error) {
	configData, err := ioutil.ReadFile(confFilePath)
	if err != nil {
		return
	}
	configDataJSON := []byte(configData)

	err = json.Unmarshal(configDataJSON, &ddnsConf)
	if err != nil {
		return
	}
	// fmt.Println(ddnsConf)
	return
}
