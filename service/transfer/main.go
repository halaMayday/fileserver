package main

import (
	"bufio"
	"encoding/json"
	"filestore-server/config"
	dblayer "filestore-server/db"
	"filestore-server/mq"
	"filestore-server/store/oss"
	"log"
	"os"
)

//ProcessTransfer:处理文件转移
func ProcessTransfer(msg []byte) bool {
	log.Println(string(msg))

	pubData := mq.TransferData{}
	//将json转化为消息结构体
	err := json.Unmarshal(msg, &pubData)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	//打开文件
	fin, err := os.Open(pubData.CurLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	//上传到oss
	err = oss.Bucket().PutObject(
		pubData.DestLocation,
		bufio.NewReader(fin))
	if err != nil {
		log.Println(err.Error())
		return false
	}
	//更新文件转移后的地址
	_ = dblayer.UpdateFileLocation(
		pubData.FileHash,
		pubData.DestLocation)
	return true
}

func main() {
	if !config.AsyncTransferEnable {
		log.Println("异步转移文件功能目前被禁用，请检查相关配置")
		return
	}
	log.Println("文件转移服务启动中，开始监听转移任务队列...")
	mq.StartConsume(
		config.TransOSSErrQueueName,
		"transer_oss",
		ProcessTransfer)
}
