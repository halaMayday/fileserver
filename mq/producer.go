package mq

import (
	"filestore-server/config"
	"log"

	"github.com/streadway/amqp"
)

var conn *amqp.Connection
var channel *amqp.Channel

//如果异常关闭,会接收到通知
var notifyClose chan *amqp.Error

func init() {
	//是否开启异步转移功能，开启的时候才初始化RabbitMQ连接
	if !config.AsyncTransferEnable {
		return
	}
	if initChannel() {
		channel.NotifyClose(notifyClose)
	}
	//断线自动重连
	go func() {
		for {
			select {
			case msg := <-notifyClose:
				conn = nil
				channel = nil
				log.Panicf("onNotifyChannelClosed: %+v\n", msg)
				initChannel()
			}
		}
	}()
}

func initChannel() bool {
	//1.判断channel是否已经创建过连接
	if channel != nil {
		return true
	}
	//2.获取一个RabbitMQ连接
	conn, err := amqp.Dial(config.RabbitURL)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	//3.打开一个channel,用于消息的发布与接收
	channel, err = conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}

//Publish:发布消息
func Publish(exchange, routeKey string, msg []byte) bool {
	if !initChannel() {
		return false
	}

	if nil == channel.Publish(
		exchange,
		routeKey,
		false, //如果没有对应的queue，就会丢弃这条消息
		false,
		amqp.Publishing{
			ContentType: "text/palin",
			Body:        msg}) {
		return true
	}
	return false
}
