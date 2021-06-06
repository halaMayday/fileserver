package ceph

import (
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

var cephConn *s3.S3

//获取Ceph链接
func GetCephConnection() *s3.S3 {
	if cephConn != nil {
		return cephConn
	}
	//1.初始化ceph的一些信息。todo:需要替换
	auth := aws.Auth{
		AccessKey: "123",
		SecretKey: "123",
	}
	//2.创建S3类型的连接
	curRegion := aws.Region{
		Name:                 "default",
		EC2Endpoint:          "http:127.0.0.1:9080",
		S3Endpoint:           "http:127.0.0.1:9080",
		S3BucketEndpoint:     "",
		S3LocationConstraint: false,
		S3LowercaseBucket:    false,
		Sign:                 aws.SignV2,
	}
	return s3.New(auth, curRegion)
}

//获取bucket
func GetCephBucket(bucket string) *s3.Bucket {
	conn := GetCephConnection()
	return conn.Bucket(bucket)
}
