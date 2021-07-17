package config

import "fmt"

var (
	MySQLSource = "root:hufan1qaz@WSX@tcp(192.168.0.112)/fileserver?charset=utf8"
)

func UpdateDBHost(host string) {
	MySQLSource = fmt.Sprintf("root:hufan1qaz@WSX@tcp(%s)/fileserver?charset=utf8", host)
}
