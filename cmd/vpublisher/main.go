package main

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"github.com/zppro/go-common/file"
	"github.com/zppro/vchecker/internal/pkg/shared"
	"io"
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
	f0 := file.GetAllFile(path, patterns)
	fmt.Println(f0)
	files := GetAllFile(path, patterns)
	for _, v := range files {
		fmt.Println(v)
		parts := strings.Split(v, ",")
		var filename, appId, output, biz, ver, size, hash string
		switch len(parts) {
		case 4:
			filename, appId, output, biz, ver, size, hash = parts[0], parts[1], "jssdk", "all", "dev", parts[2], parts[3]
			break
		case 6:
			filename, appId, output, biz, ver, size, hash  = parts[0], parts[1], parts[2], "all", parts[3], parts[4], parts[5]
			break
		case 7:
			filename, appId, output, biz, ver , size, hash = parts[0], parts[1], parts[2], parts[3], parts[4], parts[5], parts[6]
			break
		}
		if len(appId) > 0 {
			Save2JSON("1.json", filename, appId, output, biz, ver , size, hash)
			fmt.Printf("filename:%s =>{appId:%s, output:%s, ver:%s, biz:%s, size:%s, hash:%s}\n", filename, appId, output, ver, biz, size, hash)
		}
	}
	fmt.Println("出口")
}

func Save2JSON (savePath, filename, appId, output, biz, ver, size, hash string) {
	for _, v := range shared.Stages {
		fmt.Println(v.Domain, savePath, filename, appId, output, biz, ver, size, hash)
	}
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
					files = append(files, fmt.Sprintf("%s,%s,%d,%s", filename, reg.ReplaceAllString(filename, "$1"), fi.Size(), GetFileHash(path + "/" + filename)))
					//fmt.Printf("dev file %q, %d\n", files, len(files))
				}
			}
		}
	}
	return
}

func GetFileHash(filename string) (hash string) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
		return
	}
	defer f.Close()

	br := bufio.NewReader(f)

	h := sha1.New()
	_, err = io.Copy(h, br)

	if err != nil {
		panic(err)
		return
	}
	hash = fmt.Sprintf("%x", h.Sum(nil))
	return
}