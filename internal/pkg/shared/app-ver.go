package shared

type AppVerData struct {
	Url   string `json:"url"`
	Hash  string `json:"hash"`
}

type AppVer struct {
	AppId string `json:"appId"`
	AppOutput string `json:"output"`
	Biz string `json:"biz"`
	Stage string `json:"stage"`
	Ver   string `json:"ver"`
	Data AppVerData `json:"data"`
}

type AppVersions []AppVer

func (vers *AppVersions) Filter(predict filterFunc) []AppVer {
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

func (vers *AppVersions) Find(predict findFunc)(found *AppVer) {
	for _, v := range *vers {
		if match := predict(v); match {
			found = &v
			break
		}
	}

	return
}