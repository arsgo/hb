package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"
	"time"
)

var paramsChan chan map[string]interface{}

func init() {
	paramsChan = make(chan map[string]interface{}, 100000)
}

func (c *httpClient) ResultChanHanlde(content []byte) (b bool, err error) {
	b = len(content) != 0
	if !b {
		return
	}
	defer func() {
		b = isSuccess(content)
	}()
	input := make(map[string]interface{})
	if strings.HasPrefix(string(content), "<?") {
		err = xml.Unmarshal(content, &input)
	} else {
		err = json.Unmarshal(content, &input)
	}
	if err != nil {
		return
	}
	fields := c.data.Params["->"]
	if fields != "*" {
		data := input[fields].(map[string]interface{})
		paramsChan <- data
	} else {
		paramsChan <- input
	}
	return
}

func (c *httpClient) getFromChan() (r map[string]interface{}, er error) {
	if !strings.EqualFold(c.data.Params["<-"], "*") {
		r = make(map[string]interface{})
		return
	}
	ticker := time.NewTicker(time.Second * 10)
	select {
	case r = <-paramsChan:
	case <-ticker.C:
		{
			er = errors.New("get data timeout")
			r = make(map[string]interface{})
			return
		}
	}
	return

}
