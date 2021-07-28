package mq

import (
	"filestore-server/config"
	"github.com/streadway/amqp"
	"log"
)

//如果异关闭，会接收通知
var notifyClose chan *amqp.Error

//UploadRabbitHost：更新mq host
func UpdateRabbitHost(host string) {
	config.RabbitURL = host
}

// Init : 初始化MQ连接信息
func Init() {
	// 是否开启异步转移功能，开启时才初始化rabbitMQ连接
	if !config.AsyncTransferEnable {
		return
	}
	if initChannel(config.RabbitURL) {
		channel.NotifyClose(notifyClose)
	}
	// 断线自动重连
	go func() {
		for {
			select {
			case msg := <-notifyClose:
				conn = nil
				channel = nil
				log.Printf("onNotifyChannelClosed: %+v\n", msg)
				initChannel(config.RabbitURL)
			}
		}
	}()
}

func initChannel(rabbitHost string) bool {
	if channel != nil {
		return true
	}

	conn, err := amqp.Dial(rabbitHost)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	channel, err = conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}
