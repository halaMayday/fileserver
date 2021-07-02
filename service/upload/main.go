package main

import (
	cfg "filestore-server/config"
	"filestore-server/route"
	"fmt"
)

func main() {
	router := route.Router()
	fmt.Printf("上传服务启动中，开始监听监听[%s]...\n", cfg.UploadServiceHost)
	router.Run(cfg.UploadServiceHost)
}
