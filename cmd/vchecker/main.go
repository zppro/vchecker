package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"github.com/zppro/vchecker/internal/pkg/shared"
	"github.com/zppro/vchecker/internal/pkg/vchecker"
)


var mktVers shared.AppVersions
var cache *vchecker.FilterCache

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

	key := fmt.Sprintf("%s_%s", appId, ver)
	value, ok := cache.Get(key)
	if !ok {
		value = mktVers.Find(func(item shared.AppVer) bool {
			return item.Ver == ver
		})
		cache.Set(key, value)
	}

	toJson(w, value)
}

func toJson (w http.ResponseWriter, item *shared.AppVer) {
	if nil == item {
		w.WriteHeader(404)
		return
	}
	data, err := json.Marshal(item)
	if err != nil {
		log.Printf("JSON marshaling faild: %s\n", err)
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Write(data)
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
	cache = vchecker.NewFilterCache()
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerDefault)
	mux.HandleFunc("/check", handlerCheck)
	http.ListenAndServe(":8083", mux)
}
