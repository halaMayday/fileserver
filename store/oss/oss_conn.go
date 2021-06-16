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
