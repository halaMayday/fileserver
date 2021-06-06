package main

import (
	"filestore-server/store/ceph"
	"fmt"
	"os"

	"gopkg.in/amz.v1/s3"
)

func main() {
	//获取bucket
	bucket := ceph.GetCephBucket("testbucket1")

	//创建一个新的bucket(PublicRead 表示所有用户都可以访问这个bucket)
	err := bucket.PutBucket(s3.PublicRead)
	fmt.Printf("create bucket err:%v\n", err)

	//查询这个bucket下面指定条件的object keys
	objectRes, err := bucket.List("", "", "", 100)
	fmt.Printf("object keys:%+v\n", objectRes)

	//新上传一个对象
	err = bucket.Put("/testupload/a.txt", []byte("just for test"), "octet-stream", s3.PublicRead)
	fmt.Printf("upload err:%v\n", err)

	//查询这个bucket下面指定条件的object keys
	objectRes, err = bucket.List("", "", "", 100)
	fmt.Printf("object keys:%+v\n", objectRes)

	//测试从ceph中获取数据，写入到本地
	fileBytes, err := bucket.Get("ceph/xxxxxx")
	tmpFile, _ := os.Create("/tmp/test_file")
	tmpFile.Write(fileBytes)
}
