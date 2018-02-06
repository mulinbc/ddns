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

	err = ddns.init()
	if err != nil {
		log.Fatal(err)
	}

	ipChan := make(chan string)
	go ddns.GetIP.getLocalIP(ipChan)

	oldIP := ""
	newIP := ""
	for {
		newIP = <-ipChan
		if newIP != oldIP {
			//Update DNS record
			ddns.updateDNSRecord(newIP)
			//Send email
			ddns.Mail.Content = "DNS record has been updated:\r\n"
			for _, u := range ddns.DNSRecord {
				ddns.Mail.Content += u.Name + ": " + newIP + "\r\n"
			}
			go ddns.Mail.sendEmail()

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
	return
}
