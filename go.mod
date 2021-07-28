module filestore-server

go 1.14

require (
	github.com/aliyun/aliyun-oss-go-sdk v0.0.0-20190307165228-86c17b95fcd5
	github.com/garyburd/redigo v1.6.0
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.6.2
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.4.0
	github.com/json-iterator/go v1.1.9
	github.com/micro/cli v0.2.0
	github.com/micro/go-micro v1.18.0
	github.com/micro/go-plugins/registry/consul v0.0.0-20200119172437-4fe21aa238fd
	github.com/mitchellh/mapstructure v1.1.2
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/stretchr/testify v1.5.1 // indirect
	golang.org/x/net v0.0.0-20191109021931-daa7c04131f5
	google.golang.org/protobuf v1.21.0
	gopkg.in/amz.v1 v1.0.0-20150111123259-ad23e96a31d2
)

replace (
	github.com/micro/go-micro v1.18.0 => github.com/micro/go-micro v1.18.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)
