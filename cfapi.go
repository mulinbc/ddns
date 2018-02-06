package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const endpoints = "https://api.cloudflare.com/client/v4/"
const contentType = "application/json"

type secret struct {
	XAuthEmail string `json:"x_auth_email"`
	XAuthKey   string `json:"x_auth_key"`
}

type dnsRecord struct {
	id      string
	Type    string `json:"type"` //valid values: A, AAAA, CNAME, TXT, SRV, LOC, MX, NS, SPF
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	zoneID  string
}

type ddns struct {
	GetIP     getIP         `json:"get_ip"`
	Secret    secret        `json:"secret"`
	DNSRecord []dnsRecord   `json:"dns_record"`
	Mail      mail          `json:"mail"`
	Duration  time.Duration `json:"duration"`
}

type dnsRecordResponse struct {
	DNSRecordID string `json:"id"`
	Name        string `json:"name"`
}
type dnsListResponse struct {
	Result []dnsRecordResponse `json:"result"`
}
type dnsCreateResponse struct {
	Result dnsRecordResponse `json:"result"`
}

type zoneResponse struct {
	ZoneID string `json:"id"`
	Name   string `json:"name"`
}
type zoneListResponse struct {
	Result []zoneResponse `json:"result"`
}

//Only support update A record
func (p *ddns) updateDNSRecord(ip string) (err error) {
	for _, u := range p.DNSRecord {
		err := p.updateOneDNSRecord(ip, u)
		if err != nil {
			for {
				log.Println("Update dns failed! Retrying...")
				err := p.updateOneDNSRecord(ip, u)
				if err == nil {
					break
				}
				time.Sleep(p.Duration)
			}
		}
	}
	return err
}

func (p *ddns) init() (err error) {

	oldMainDomain := ""
	newMainDomain := ""
	zoneID := ""
	dnsID := ""

	for i := range p.DNSRecord {
		newMainDomain, err = splitMainDomain(p.DNSRecord[i].Name)
		if err != nil {
			// log.Fatal(err)
			return
		}
		if newMainDomain != oldMainDomain {
			oldMainDomain = newMainDomain
			zoneID, err = p.getZoneIDbyName(newMainDomain)
			if err != nil {
				return
			}
			p.DNSRecord[i].zoneID = zoneID

			dnsID, err = p.getDNSIDbyName(p.DNSRecord[i].Name, zoneID)
			if err != nil {
				dnsID, err = p.createOneDNSRecord(p.DNSRecord[i].Content, p.DNSRecord[i])
				if err != nil {
					return
				}
			}
			p.DNSRecord[i].id = dnsID
		} else {
			p.DNSRecord[i].zoneID = zoneID
			dnsID, err = p.getDNSIDbyName(p.DNSRecord[i].Name, zoneID)
			if err != nil {
				dnsID, err = p.createOneDNSRecord(p.DNSRecord[i].Content, p.DNSRecord[i])
				if err != nil {
					return
				}
			}
			p.DNSRecord[i].id = dnsID
		}
	}
	return
}

func (p *ddns) getZoneIDbyName(name string) (zoneID string, err error) {
	count := 0
	for {
		zoneList, err := p.listZone()
		if err == nil {
			for _, t := range zoneList.Result {
				if t.Name == name {
					zoneID = t.ZoneID
					break
				}
				count++
			}
			if count == len(zoneList.Result) {
				err = errors.New(name + " is not in zone, add it")
				// log.Println(err)
				return "", err
			}
			break
		}
		time.Sleep(p.Duration)
	}
	return
}

func (p *ddns) listZone() (zoneIDandName zoneListResponse, err error) {

	client := &http.Client{}
	url := endpoints + "zones"

	request, err := http.NewRequest(http.MethodGet, url, strings.NewReader(""))
	if err != nil {
		log.Println(err)
		return
	}
	defer request.Body.Close()
	request.Header.Set("X-Auth-Email", p.Secret.XAuthEmail)
	request.Header.Set("X-Auth-Key", p.Secret.XAuthKey)
	request.Header.Set("Content-Type", contentType)

	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return
	}
	defer response.Body.Close()

	resbody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return
	}
	// fmt.Println(string(resbody))

	err = json.Unmarshal(resbody, &zoneIDandName)
	if err != nil {
		log.Println(err)
		return
	}
	return
}

func (p *ddns) getDNSIDbyName(name string, zoneID string) (dnsRecordID string, err error) {
	count := 0
	for {
		dnsRecordList, err := p.listDNSRecord(zoneID)
		if err == nil {
			for _, t := range dnsRecordList.Result {
				if t.Name == name {
					dnsRecordID = t.DNSRecordID
					break
				}
				count++
			}
			if count == len(dnsRecordList.Result) {
				err = errors.New(name + " is not in DNS record, add it")
				// log.Println(err)
				return "", err
			}
			break
		}
		time.Sleep(p.Duration)
	}
	return
}

//return zoneID&name
func (p *ddns) listDNSRecord(zoneID string) (dnsListResponse dnsListResponse, err error) {
	client := &http.Client{}
	url := endpoints + "zones/" + zoneID + "/dns_records"

	request, err := http.NewRequest(http.MethodGet, url, strings.NewReader(""))
	if err != nil {
		log.Println(err)
		return
	}
	defer request.Body.Close()
	request.Header.Set("X-Auth-Email", p.Secret.XAuthEmail)
	request.Header.Set("X-Auth-Key", p.Secret.XAuthKey)
	request.Header.Set("Content-Type", contentType)

	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return
	}
	defer response.Body.Close()

	resbody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return
	}
	// fmt.Println(string(resbody))

	err = json.Unmarshal(resbody, &dnsListResponse)
	if err != nil {
		log.Println(err)
		return
	}
	return
}

func (p *ddns) createOneDNSRecord(ip string, dnsRecord dnsRecord) (dnsID string, err error) {

	dnsRecord.Content = ip
	dRJson, err := json.Marshal(dnsRecord)
	if err != nil {
		log.Println(err)
		return
	}

	client := &http.Client{}
	url := endpoints + "zones/" + dnsRecord.zoneID + "/dns_records"
	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(dRJson)))
	if err != nil {
		log.Println(err)
		return
	}
	defer request.Body.Close()
	request.Header.Set("X-Auth-Email", p.Secret.XAuthEmail)
	request.Header.Set("X-Auth-Key", p.Secret.XAuthKey)
	request.Header.Set("Content-Type", contentType)

	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return
	}
	defer response.Body.Close()

	resbody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return
	}

	// fmt.Println(string(resbody))
	log.Println("Create DNS record: " + dnsRecord.Name + ": " + ip)

	var dnsCreateResponse dnsCreateResponse
	err = json.Unmarshal(resbody, &dnsCreateResponse)
	if err != nil {
		log.Println(err)
		return
	}
	dnsID = dnsCreateResponse.Result.DNSRecordID
	return
}

func (p *ddns) updateOneDNSRecord(ip string, dnsRecord dnsRecord) (err error) {

	dnsRecord.Content = ip
	dRJson, err := json.Marshal(dnsRecord)
	if err != nil {
		log.Println(err)
		return
	}

	client := &http.Client{}
	url := endpoints + "zones/" + dnsRecord.zoneID + "/dns_records/" + dnsRecord.id
	request, err := http.NewRequest(http.MethodPut, url, strings.NewReader(string(dRJson)))
	if err != nil {
		log.Println(err)
		return
	}
	defer request.Body.Close()
	request.Header.Set("X-Auth-Email", p.Secret.XAuthEmail)
	request.Header.Set("X-Auth-Key", p.Secret.XAuthKey)
	request.Header.Set("Content-Type", contentType)

	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return
	}
	defer response.Body.Close()

	// resbody, err := ioutil.ReadAll(response.Body)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// fmt.Println(string(resbody))
	log.Println("Update DNS record: " + dnsRecord.Name + ": " + ip)

	return
}

func (p *ddns) deleteOneDNSRecord(zoneID string, dnsRecordID string) (err error) {

	client := &http.Client{}
	url := endpoints + "zones/" + zoneID + "/dns_records/" + dnsRecordID
	request, err := http.NewRequest(http.MethodDelete, url, strings.NewReader(""))
	if err != nil {
		log.Println(err)
		return
	}
	defer request.Body.Close()

	request.Header.Set("X-Auth-Email", p.Secret.XAuthEmail)
	request.Header.Set("X-Auth-Key", p.Secret.XAuthKey)
	request.Header.Set("Content-Type", contentType)

	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return
	}
	defer response.Body.Close()

	resbody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return
	}

	//TODO 对删除进行确认
	fmt.Println(string(resbody))

	return
}

func splitMainDomain(raw string) (new string, err error) {
	s := strings.Split(raw, ".")
	// fmt.Println(s)
	len := len(s)
	if len < 2 {
		err = errors.New("Domain is wrong")
		return
	}
	new = s[len-2] + "." + s[len-1]
	return
}
