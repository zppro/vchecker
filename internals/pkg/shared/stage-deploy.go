package shared

import "fmt"

type StageDeploy struct {
	Name string
	Protocol string
	Domain string
	DownloadFragment string
}

type StageDeploys []StageDeploy

var Stages = StageDeploys{
	{"test",  "https", "mmo.downtown8.cn", "/docs/d/"},
	{"prod",  "https", "mmo.eshine.cn", "/docs/d/"},
}// "test", "bts", "prod"

func (sd *StageDeploy) GetResourceUrl (output, name string) string {
	return fmt.Sprintf("%s://%s%s%s/%s", sd.Protocol, sd.Domain, sd.DownloadFragment, output, name)
}

func (sds *StageDeploys) Find(predict findFunc)(found *StageDeploy) {
	for _, v := range *sds {
		if match := predict(v); match {
			found = &v
			break
		}
	}
	return
}