package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type AppVer struct {
	AppId string `json:"appId"`
	Ver   string `json:"ver"`
	Url   string `json:"url"`
	Hash  string `json:"hash"`
}
type MKTVersions []AppVer

type filterFunc func(AppVer) bool
type findFunc func(AppVer) bool


var mktVers MKTVersions

func (vers *MKTVersions) filter(predict filterFunc) []AppVer {
	filtered := make([]AppVer, len(*vers))
	var last int
	for i, v := range *vers {
		if match := predict(v); match {
			filtered[i] = v
			last++
		}
	}
	return filtered[0:last]
}

func (vers *MKTVersions) find(predict findFunc)(found *AppVer) {
	for _, v := range *vers {
		if match := predict(v); match {
			found = &v
			break
		}
	}

	return
}

func handlerDefault(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello => %s", r.URL.Path)
}

func handlerCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		fmt.Fprint(w, "need GET")
		return
	}
	appId := r.URL.Query().Get("appId")
	if appId == "" {
		appId = "mkt"
	}

	ver := r.URL.Query().Get("ver")
	if ver == "" {
		ver = "latest"
	}


	/// todo 考虑增加 mkt+ver的filter结果缓存并使其并发安全

	found := mktVers.find(func(item AppVer) bool {
		return item.Ver == ver
	})

	if nil == found {
		w.WriteHeader(404)
		return
	}
	data, err := json.Marshal(found)
	if err != nil {
		log.Fatalf("JSON marshaling faild: %s", err)
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Write(data)
	//fmt.Fprint(w, data)
}

func isFileExist(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		fmt.Println(info)
		return false
	}
	return true
}

func readAppVersFromFile(appId string) {
	filename := fmt.Sprintf("./%s-ver.json", appId)
	if !isFileExist(filename) {
		log.Fatalf("文件%s 不存在", filename)
		return
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("读取文件%s faild: %s", filename, err)
		return
	}
	//读取的数据为json格式，需要进行解码
	err = json.Unmarshal(data, &mktVers)
	if err != nil {
		log.Fatalf("JSON Unmarshal faild: %s", err)
		return
	}
}

func main() {
	readAppVersFromFile("mkt")
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerDefault)
	mux.HandleFunc("/check", handlerCheck)
	http.ListenAndServe(":8081", mux)
}
