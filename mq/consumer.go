package mq

import "log"

var done chan bool

func StartConsume(queueName, consumerName string, callback func(msg []byte) bool) {
	msgs, err := channel.Consume(
		queueName,
		consumerName,
		true,  //自动应答
		false, //非唯一消费者
		false, //rabbitMQ只能设置为false
		false, //nowait.false表示会阻塞知道消息过来
		nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	done = make(chan bool)

	go func() {
		//循环获取channel中的数据
		for d := range msgs {
			processErr := callback(d.Body)
			if processErr {
				//todo:将任务写入到错误队列，等待后续的处理
				log.Println("todo:需要将信息传递到错误队列中去，等待后续的处理 ")
			}
		}
	}()

	//接收done的信号，没有信息传递则会一直阻塞，避免该函数退出
	<-done

	//关闭通道
	channel.Close()
}

//StopConsume:停止监听队列
func StopConsume() {
	done <- true
}
