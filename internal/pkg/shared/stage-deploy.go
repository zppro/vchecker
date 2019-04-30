package shared

type StageDeploy struct {
	Name string
	Protocol string
	Domain string
	DownloadFragment string
}

var Stages = []StageDeploy{
	{"test",  "https", "mmo.downtown8.cn", "/docs/#/d/"},
	{"prod",  "https", "mmo.eshine.cn", "/docs/#/d/"},
}// "test", "bts", "prod"
