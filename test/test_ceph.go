package main

import (
	cfg "filestore-server/config"
	"filestore-server/store/ceph"
	"fmt"
	"os"
)

func main() {
	//获取bucket
	bucket := ceph.GetCephBucket(cfg.OSSBucket)

	//创建一个新的bucket(PublicRead 表示所有用户都可以访问这个bucket)
	//err := bucket.PutBucket(s3.PublicRead)
	//fmt.Printf("create bucket err:%v\n", err)
	//
	//查询这个bucket下面指定条件的object keys
	//objectRes, err := bucket.List("", "", "", 100)
	//fmt.Printf("object keys:%+v\n", objectRes)

	//新上传一个对象
	err := ceph.PutObject(cfg.OSSBucket, "/ceph/test1.txt", []byte("just for test"))
	fmt.Printf("upload err:%v\n", err)

	//查询这个bucket下面指定条件的object keys
	//objectRes, err = bucket.List("", "", "", 100)
	//fmt.Printf("object keys:%+v\n", objectRes)

	//测试从ceph中获取数据，写入到本地
	fileBytes, err := bucket.Get("ceph/test1.txt")
	tmpFile, _ := os.Create("/Users/nuc/gitub_space/fileserver/test/a.txt")
	tmpFile.Write(fileBytes)
}
