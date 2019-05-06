package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/zppro/go-common/file"
	"github.com/zppro/vchecker/internals/pkg/shared"
	"github.com/zppro/vchecker/internals/pkg/vchecker"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"
)

type GenerateParam struct {
	AppIds []string `json:"appIds"`
	Env string `json:"env"`
}

//var mktVers shared.AppVersions
var appVerMap map[string]shared.AppVersions
var cache *vchecker.FilterCache

func handleDefault(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello => %s", r.URL.Path)
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "need POST", 405)
		return
	}
	var p GenerateParam
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	value := shared.Stages.Find(func(item interface{}) bool {
		if v, ok := item.(shared.StageDeploy); ok {
			return v.Name == p.Env
		}
		return false
	})

	if nil == value {
		http.Error(w, "无效的env", 400)
		return
	}
	signals := make(map[string] chan struct{})
	blockSize := 1024 * 32
	for _, appId := range p.AppIds {
		key := value.GetResourceUrl("assets", appId + ".json")
		signals[key] = make(chan struct{})

		go func (downloadUrl string, signal chan struct{}) {
			log.Printf("downloadUrl:%s\n", downloadUrl)
			defer close(signal)

			uri, err := url.ParseRequestURI(downloadUrl)
			if err != nil {
				http.Error(w, "下载地址错误", 400)
				return
			}
			filename := path.Base(uri.Path)
			client := http.DefaultClient;
			client.Timeout = time.Second * 30 //设置超时时间
			resp, err := client.Get(downloadUrl)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if resp.StatusCode != 200 {
				http.Error(w, "[*] 请求文件异常.", 400)
				return
			}
			if resp.ContentLength <= 0 {
				http.Error(w, "[*] 空文件.", 400)
				return
			}
			raw := resp.Body
			defer raw.Close()
			reader := bufio.NewReaderSize(raw, blockSize)
			filepath := fmt.Sprintf("./tmp/")
			err = os.MkdirAll(filepath, os.ModePerm)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			f, err := os.Create(filepath+filename)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			writer := bufio.NewWriter(f)

			buff := make([]byte, blockSize)
			written := 0
			for {
				nr, er := reader.Read(buff)
				//log.Printf("read buff:%d, %v\n", nr, er)
				if nr > 0 {
					nw, ew := writer.Write(buff[0:nr])
					if nw > 0 {
						written += nw
					}
					if ew != nil {
						err = ew
						break
					}
					if nr != nw {
						err = io.ErrShortWrite
						break
					}
				}
				if er != nil {
					if er != io.EOF {
						err = er
					}
					break
				}
			}
			//log.Printf("for end writed: %d %v\n", written, err)

			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			writer.Flush()
		}(key, signals[key])
	}

	// 阻塞下载goroutine
	for _, v := range signals {
		<- v
	}

	log.Printf("下载执行完成!\n" )
	patterns := []string{`(?U)([a-z]+)\.json$`}
	fileExtInfos := file.GetAllFileExtInfo("./tmp", patterns)
	destPath := fmt.Sprintf("./gen/")
	err = os.MkdirAll(destPath, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	for _, v := range fileExtInfos {
		source := v.FullName()

		// 拷贝保护，预先检测下载文件是否可用
		data, err := ioutil.ReadFile(source)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		var ptr = new(shared.AppVersions)
		//读取的数据为json格式，需要进行解码
		err = json.Unmarshal(data, ptr)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		// 能够解码证明文件没有问题，可以拷贝
		log.Println("【验证】能够解码证明文件没有问题，可以拷贝!")
		dest := destPath + v.Name()
		log.Printf("【拷贝】从%s到%s!\n", source, dest)
		_ , err = file.CopyFile(source, dest)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err = os.Remove(source)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	}
	log.Printf("拷贝完成!\n" )

	readAppVersFromFile()
	//w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Write([]byte("ok"))
}



func handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		fmt.Fprint(w, "need GET")
		return
	}
	appId := r.URL.Query().Get("appId")
	if appId == "" {
		appId = "mkt"
	}
	var appVersions = appVerMap[appId]
	if len(appVersions) == 0 {
		w.WriteHeader(400)
		fmt.Fprint(w, "无效的appId")
		return
	}
	biz := r.URL.Query().Get("biz")
	if biz == "" {
		biz = "all"
	}
	stage := r.URL.Query().Get("stage")
	if stage == "" {
		stage = "prod"
	}
	ver := r.URL.Query().Get("ver")
	if ver == "" {
		ver = "latest"
	}

	key := fmt.Sprintf("%s_%s_%s_%s", appId, biz, stage, ver)
	value, ok := cache.Get(key)
	if !ok {
		value = appVersions.Find(func(item interface{}) bool {
			if v, ok := item.(shared.AppVer); ok {
				return v.Biz == biz && v.Stage == stage && v.Ver == ver
			}
			return false
		})
		if value != nil {
			cache.Set(key, value)
		}
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

func readAppVersFromFile() {
	patterns := []string{`(?U)([a-z]+)\.json$`}
	fileExtInfos := file.GetAllFileExtInfo("./gen", patterns)
	appVerMap = make(map[string]shared.AppVersions, len(fileExtInfos))
	for _, v := range fileExtInfos {
		filename := v.FullName()
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("读取文件%s faild: %s", filename, err)
			return
		}

		var ptr = new(shared.AppVersions)
		//读取的数据为json格式，需要进行解码
		err = json.Unmarshal(data, ptr)
		if err != nil {
			log.Fatalf("JSON Unmarshal faild: %s", err)
			return
		}
		key := v.Reg.ReplaceAllString(v.Name(), "$1")
		appVerMap[key] = *ptr
	}
}

func main() {
	readAppVersFromFile()
	cache = vchecker.NewFilterCache()
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleDefault)
	mux.HandleFunc("/check", handleCheck)
	mux.HandleFunc("/generate", handleGenerate)
	http.ListenAndServe(":8083", mux)
}
