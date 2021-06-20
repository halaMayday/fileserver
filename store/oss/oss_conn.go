package oss

import (
	cfg "filestore-server/config"
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var ossCli *oss.Client

//Client:创建oss client 对象
func Client() *oss.Client {
	if ossCli != nil {
		return ossCli
	}

	ossCli, err := oss.New(
		cfg.OSSEndpoint,
		cfg.OSSAccesskeyID,
		cfg.OSSAccessKeySecret)

	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return ossCli
}

//Bucket:获取bucket储存空间
func Bucket() *oss.Bucket {
	cli := Client()
	if cli != nil {
		bucket, err := cli.Bucket(cfg.OSSBucket)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		return bucket
	}
	return nil
}

//downloadURL:临时授权下载文件
func DownloadURL(objName string) string {
	signURL, err := Bucket().SignURL(objName, oss.HTTPGet, 3600)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return signURL
}

//BuildLifeCycleRule:针对指定的bucker设置生命周期管理
func BuildLifeCycleRule(buckerName string) {
	//表示前缀为test的对象(文件)距最后修改时间30天后过期
	ruleTest1 := oss.BuildLifecycleRuleByDays("rule1", "/test/", true, 30)
	rules := []oss.LifecycleRule{ruleTest1}
	Client().SetBucketLifecycle(buckerName, rules)
}
