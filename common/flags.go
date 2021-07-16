package common

import "github.com/micro/cli"

//CustomFlags:自定义命令行参数
var CustomFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "dbhost",
		Value: "192.168.0.112",
		Usage: "database address",
	},
	cli.StringFlag{
		Name:  "mqhost",
		Value: "192.168.0.112",
		Usage: "mq(rabbitmq) address",
	},
	cli.StringFlag{
		Name:  "cachehost",
		Value: "192.168.0.112",
		Usage: "cache(redis) address",
	},
	cli.StringFlag{
		Name:  "cephhost",
		Value: "192.168.0.112",
		Usage: "ceph address",
	},
}
