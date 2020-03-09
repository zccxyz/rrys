package code

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

const (
	host  = "http://pc.zmzapi.com/index.php?g=api/pv3&m=index&client=5&accesskey=519f9cab85c8059d17544947k361a827&"
	page  = 1
	limit = 30
	order = "itemupdate"
)

//爬取电视剧
func TvRun() {
	body, err := GetDetail(11057)
	if err != nil {
		log.Fatal(err)
	}
	rs, err := ioutil.ReadAll(body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(rs))
}

//爬取电影
func MvRun() {
	GetDetail(11057)
}

func GetData(page int, channel string) (io.ReadCloser, error) {
	url := host + "order=" + order + "&a=resource_storage&limit=" + strconv.Itoa(limit) + "&page=" + strconv.Itoa(page) + "&channel=" + channel
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func GetDetail(vid int) (io.ReadCloser, error) {
	url := host + "a=resource&id=" + strconv.Itoa(vid)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
