package config

const (
	//AsyncTransferEnable:是否开启文件异步转移 （默认开启同步）
	AsyncTransferEnable = true
	//RabbitURL:RabbitMQ服务入口的url
	RabbitURL = "amqp://guest:guest@192.168.0.112:5672/"
	//TransExchangeName:用于文件transfer的交换机
	TransExchangeName = "uploadserver.trans"
	//TransOSSQueueName:oss转移队列名
	TransOSSQueueName = "uploadserver.trans.oss"
	//TransOSSErrQueueName:oss转移失败后的写入另外一个队列的队列名
	TransOSSErrQueueName = "uploadserver.trans.oss.err"
	//TransOSSroutingKey:routingkey
	TransOSSroutingKey = "oss"
)
