package main

import (
	"encoding/json"
	"fmt"
	"github.com/zppro/go-common/file"
	"github.com/zppro/vchecker/internal/pkg/shared"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func main() {
	path := os.Args[0]
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	fmt.Println(path)
	patterns := []string{`(?U)([a-z]+)\.dev\.js$`, `(?U)([0-9a-z\-\.]+)\.min\.js$`}
	//f0 := file.GetAllFile(path, patterns)
	//fmt.Println(f0)
	fileExtInfos := file.GetAllFileExtInfo(path, patterns)
	var appVerMap = make(map[string]shared.AppVersions, 2)

	for _, v := range fileExtInfos {
		filename, size, hash := v.Name(), v.FI.Size(), v.Hash()
		//fmt.Printf("%v\n", filename)
		parts := strings.Split(v.Reg.ReplaceAllString(filename, "$1"), "-")
		var appId, output, biz, ver string
		//fmt.Printf("%v\n", parts)
		switch len(parts) {
		case 1:
			appId, output, biz, ver = parts[0], "jssdk", "all", "dev"
			break
		case 3:
			appId, output, biz, ver = parts[0], parts[1], "all", parts[2]
			break
		case 4:
			appId, output, biz, ver = parts[0], parts[1], parts[2], parts[3]
			break
		}
		//fmt.Printf("appId:%s-%s-%s-%s\n", appId, output, biz, ver)
		if len(appId) > 0 {
			if appVerMap[appId] == nil {
				appVerMap[appId] = make(shared.AppVersions, 0)
			}
			//fmt.Printf("appVerMap[appId]:%v\n", appVerMap[appId])
			for _, stage := range shared.Stages {

				appVerMap[appId] = append(appVerMap[appId], shared.AppVer{
					AppId: appId,
					AppOutput: output,
					Biz: biz,
					Stage: stage.Name,
					Ver: ver,
					Data: shared.AppVerData{
						Url: stage.GetResourceUrl(output, filename),
						Size: size,
						Hash: hash,
					},
				})
			}
		}

	}
	for k, v := range appVerMap {
		Save2JSON(fmt.Sprintf("./assets/%s.json", k), v)
	}

	fmt.Println("出口")
}

func Save2JSON (saveAs string, appVerMap shared.AppVersions) error {
	jsonBytes, err := json.MarshalIndent(appVerMap, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(saveAs, jsonBytes, 0666)
	return err
}

func GetAllFile(path string, patterns []string) (files []string) {
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Printf("ReadDir error %s\n", err)
		return
	}

	for _, fi := range entries {
		filename := fi.Name()
		if fi.IsDir() {
			fmt.Printf("dir [%s]\n", path+"/"+filename)
			files = append(files, GetAllFile(path + fi.Name() + "/", patterns)...)
		} else {
			for _, pattern := range patterns {
				if ok,_ := regexp.MatchString(pattern, filename); ok {
					reg := regexp.MustCompile(pattern)
					files = append(files, fmt.Sprintf("%s,%s,%d,%s", filename, reg.ReplaceAllString(filename, "$1"), fi.Size(), file.GetFileHash(path + "/" + filename)))
					//fmt.Printf("dev file %q, %d\n", files, len(files))
				}
			}
		}
	}
	return
}