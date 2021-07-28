package main

import (
	"filestore-server/common"
	cfg "filestore-server/config"
	"filestore-server/mq"
	"filestore-server/route"
	dbproxy "filestore-server/service/dbproxy/client"
	upRpc "filestore-server/service/upload/rpc"
	"fmt"
	"github.com/micro/cli"
	"github.com/micro/go-micro"
	"log"
	"time"
)

func main() {
	startAPIService()
}

//startAPIService:启动 api 服务
func startAPIService() {
	router := route.Router()
	fmt.Printf("上传服务启动中，开始监听监听[%s]...\n", cfg.UploadServiceHost)
	router.Run(cfg.UploadServiceHost)
}

func startRPCService() {
	service := micro.NewService(
		micro.Name("go.mirco.service.upload"),  //服务名称
		micro.RegisterInterval(time.Second*10), //TTL 指定从上一次心跳开始，超过这个时间服务就会被移除
		micro.RegisterInterval(time.Second*5),  // 让服务在指定时间内重新注册，保持TTL获取的注册时间有效
		micro.Flags(common.CustomFlags...))

	service.Init(
		micro.Action(func(context *cli.Context) {
			//检查是否指定想mqhost
			mqhost := context.String("mqhost")
			if len(mqhost) > 0 {
				log.Println("custom mq address: " + mqhost)
				mq.UpdateRabbitHost(mqhost)
			}
		}))

	//初始化dbproxy client
	dbproxy.Init(service)
	// 初始化mq client
	mq.Init()
	//TODO:
	//upProto.RegisterUploadServiceHandler(service.Server(), new(upRpc.Upload))
	//if err := service.Run(); err != nil {
	//	fmt.Println(err)
	//}
}
