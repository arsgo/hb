package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
	"github.com/colinyl/lib4go/utility"
)


type responseXML struct {
	FailedCode   string `xml:"failedCode"`
	FailedReason string `xml:"failedReason"`
}

type httpClient struct {
	client *http.Client
	data   *dataBlock
}

func NewHttpClient(data *dataBlock) *httpClient {
	return &httpClient{client: createClient(), data: data}
}

func (c *httpClient) Reqeust() *response {
    defer func() {
		if err := recover();nil!=err {
			log.Fatal(err.(error).Error())
		}
	}()
	url := fmt.Sprintf("%s?%s", c.data.URL, c.makeParams())
	startTime := time.Now()
	resp, er := c.client.Get(url)
	endTime := time.Now()
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	success := isSuccess(body)
	if !success {
		log.Errorf("url:%s,content:%s\r\n", url, string(body))
	}
	return &response{success: resp.StatusCode == 200 && er == nil && success,
		url: c.data.URL, useTime: subTime(startTime, endTime)}
}


func (h *httpClient) makeParams() string {
	var (
		rawFormat string
		keys      []string
	)
	for k := range h.data.Params {		
		if strings.HasPrefix(k, "$") {
			rawFormat = h.data.Params[k]
			continue
		} else {
		keys = append(keys, k)
		}
		
	}
	sort.Sort(sort.StringSlice(keys))
	var keyValues []string
	var urlParams []string
   nmap := getDatamap()
	for _, k := range keys {
		var value string
		if v1, ok := h.data.Params[k]; ok {
			value = nmap.Translate(v1)
		} else {
			continue
		}
		keyValues = append(keyValues, k+value)
		urlParams = append(urlParams, k+"="+value)
	}	
	dataMap := getDatamap()
	dataMap.Set("raw", strings.Join(keyValues, ""))
	fullRaw := strings.Replace(dataMap.Translate(rawFormat), " ", "", 10)
	log.Debug(fullRaw)
	urlParams = append(urlParams, "sign="+utility.Md5(fullRaw))
	return strings.Join(urlParams, "&")
}

func isSuccess(content []byte) bool {
	if len(content) == 0 {
		return false
	}
	o := responseXML{}
	xml.Unmarshal(content, &o)
	return strings.EqualFold(o.FailedCode, "000")
}

func subTime(startTime time.Time, endTime time.Time) int {
	return int(endTime.Sub(startTime).Nanoseconds() / 1000 / 1000)
}
func getDatamap() *utility.DataMap {
	baseData := utility.NewDataMap()
	baseData.Set("guid", utility.GetGUID())
	baseData.Set("timestamp", time.Now().Format("20060102150405"))
	return baseData
}

func createClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(netw, addr, 0)
				if err != nil {
					log.Fatal("timeout")
					return nil, err
				}
				return c, nil
			},
			MaxIdleConnsPerHost:   0,
			ResponseHeaderTimeout: 0,
		},
	}
}
