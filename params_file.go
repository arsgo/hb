package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"os"
	"strings"
)

type fileInputValues struct {
	data  []map[string]interface{}
	index int
}

var fileValues map[string]*fileInputValues

func init() {
	fileValues = make(map[string]*fileInputValues)
}
func readFile(path string) (lst []map[string]interface{}, err error) {
	f, _ := os.OpenFile(path, os.O_RDONLY, 0666)
	defer f.Close()
	m := bufio.NewReader(f)
	for {
		content, err := m.ReadString('\n')
		if err != nil {
			break
		}
		data := make(map[string]interface{})
		err = json.Unmarshal([]byte(content), &data)
		if err != nil {
			break
		}
		lst = append(lst, data)
	}
	return
}

func readAndGet(path string) (r map[string]interface{}, content string, er error) {
	if strings.EqualFold(path, "") {
		r = make(map[string]interface{})
		return
	}
	if _, ok := fileValues[path]; !ok {
		lst, er := readFile(path)
		if er != nil {
			return nil, "", er
		}
		val := &fileInputValues{}
		val.data = append(val.data, lst...)
		fileValues[path] = val
	}
	values := fileValues[path]
	index := values.index
	count := len(values.data)
	r = values.data[index%count]
	values.index++
	return

}

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


func (c *httpClient) ResultHanlde(content []byte) (b bool, err error) {
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
	fields := c.data.Params["->field"]
	filePath := c.data.Params["->file"]
	saveBuffer := content
	if fields != "*" {
		saveBuffer, err = json.Marshal(input[fields])
		if err != nil {
			return
		}
	}
	fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		return
	}
	fd.WriteString(string(saveBuffer) + "\n")
	fd.Close()
	return
}

func isSuccess(content []byte) bool {
	if len(content) == 0 {
		return false
	}
	if strings.HasPrefix(string(content), "<?") {
		return strings.Contains(string(content), "<failedCode>000</failedCode>")
	}
	o := responseJson{}
	err := json.Unmarshal(content, &o)
	if err != nil {
		return false
	}
	return strings.EqualFold(o.Result.Code, "success") ||
		strings.EqualFold(o.Code, "success") || strings.EqualFold(o.Code, "100")

}
