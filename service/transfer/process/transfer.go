package process

import (
	"bufio"
	"encoding/json"
	"filestore-server/mq"
	dbcli "filestore-server/service/dbproxy/client"
	"filestore-server/store/oss"
	"log"
	"os"
)

//ProcessTransfer:处理文件转移
func Transfer(msg []byte) bool {
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
	resp, err := dbcli.UpdateFileLocation(pubData.FileHash, pubData.DestLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	if !resp.Suc {
		log.Println("更新数据库异常，请检查:" + pubData.FileHash)
		return false
	}
	return true
}
