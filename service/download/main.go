package main

import (
	"filestore-server/common"
	dbproxy "filestore-server/service/dbproxy/client"
	"filestore-server/service/download/config"
	"filestore-server/service/download/route"
	"fmt"
	"github.com/micro/go-micro"
	"time"
	downRpc "filestore-server/service/download/rpc"
	downProto "filestore-server/service/download/proto"
)

func startRPCService() {
	service := micro.NewService(
		micro.Name("go.micro.service.download"),
		micro.RegisterTTL(time.Second*10),
		micro.RegisterInterval(time.Second*5),
		micro.Flags(common.CustomFlags...))

	service.Init()

	//初始化dbproxy client
	dbproxy.Init(service)

	downProto.RegisterDownloadServiceHandler(service.Server(), new(downRpc.Download))

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

func startAPIService() {
	router := route.Router()
	router.Run(config.DownloadServiceHost)
}

func main() {
	// 启动 api 服务
	go startAPIService()
	// 启动 rpc 服务
	startRPCService()
}
