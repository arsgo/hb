package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/colinyl/lib4go/utility"
)

var minSEQValue uint64 = 100000

type response struct {
	success bool
	useTime int
	url     string
	index   int
}

type responseXML struct {
	FailedCode string `xml:"failedCode"`
}
type resultJson struct {
	Code string `json:"code"`
}
type responseJson struct {
	Result resultJson `json:"result"`
	Code   string     `json:"code"`
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
		if err := recover(); nil != err {
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
	if h.data.Params == nil || len(h.data.Params) == 0 {
		return ""
	}
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
	fullParamsMap := utility.NewDataMap()
	for _, k := range keys {
		var value string
		if v1, ok := h.data.Params[k]; ok {
			value = nmap.Translate(v1)
		} else {
			continue
		}
		fullParamsMap.Set(k, value)
		keyValues = append(keyValues, k+value)
		urlParams = append(urlParams, k+"="+value)
	}

	fullParamsMap.Set("raw", strings.Join(keyValues, ""))
	fullRaw := strings.Replace(fullParamsMap.Translate(rawFormat), " ", "", -1)
	//log.Debug(fullRaw)
	urlParams = append(urlParams, "sign="+utility.Md5(fullRaw))
	return strings.Join(urlParams, "&")
}

func isSuccess(content []byte) bool {
	if len(content) == 0 {
		return false
	}
	if strings.HasPrefix(string(content), "<?") {
		return strings.Contains(string(content), "<failedCode>000</failedCode>")
	} else {
		o := responseJson{}
		err := json.Unmarshal(content, &o)
		if err != nil {
			return false
		}
		return strings.EqualFold(o.Result.Code, "success") ||
			strings.EqualFold(o.Code, "success") || strings.EqualFold(o.Code, "100")
	}
}

func subTime(startTime time.Time, endTime time.Time) int {
	return int(endTime.Sub(startTime).Nanoseconds() / 1000 / 1000)
}
func getDatamap() utility.DataMap {
	baseData := utility.NewDataMap()
	baseData.Set("guid", utility.GetGUID())
	baseData.Set("seq", fmt.Sprintf("%d", atomic.AddUint64(&minSEQValue, 1)))
	baseData.Set("timestamp", time.Now().Format("20060102150405"))
	baseData.Set("unixtime",fmt.Sprintf("%d",time.Now().Unix()))
	baseData.Set("uxmillisecond",fmt.Sprintf("%d",time.Now().Unix()*1000))
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
