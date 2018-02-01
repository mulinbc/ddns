package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"
)

type getIP struct {
	URL      []string      `json:"url"`
	Retry    int           `json:"retry"`
	Duration time.Duration `json:"duration"`
}

//如果不能获取ip地址，重试p.Retry次后不再使用该URL,直到所有的URL都不能使用后，再重新计数
func (p *getIP) getLocalIP(ipChan chan string) {
	w := make([]int, len(p.URL))
	for {
		for i, url := range p.URL {
			if w[i] < p.Retry {
				ip, err := p.getLocalIPfromURL(url)
				//某些情况会获取不到ip地址
				if err == nil && ip != "" {
					ipChan <- ip
					// fmt.Println("Get ip from: " + url)
					break
				}
				w[i]++
			}

			if allMoreThan(w, p.Retry-1) {
				for i := range w {
					w[i] = 0
				}
			}

		}
		time.Sleep(p.Duration)
	}
}

func (p *getIP) getLocalIPfromURL(url string) (ip string, err error) {
	//Send http Get request
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		// ip = ""
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New("Get " + url + ": http status: " + resp.Status)
		log.Println(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		// ip = ""
		return
	}

	//Regular Expression
	//ipReg := regexp.MustCompile("\\d+\\.\\d+\\.\\d+\\.\\d+")
	ipReg := regexp.MustCompile("((25[0-5]|2[0-4]\\d|((1\\d{2})|([1-9]?\\d)))\\.){3}(25[0-5]|2[0-4]\\d|((1\\d{2})|([1-9]?\\d)))")
	ip = ipReg.FindString(string(body))
	return
}

//如果数组中的元素都大于num，返回true
func allMoreThan(arr []int, num int) (flag bool) {
	count := 0
	for _, t := range arr {
		if t > num {
			count++
		}
	}
	if count == len(arr) {
		flag = true
	}
	return
}
