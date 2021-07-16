package main

import (
	"filestore-server/service/apigw/route"
)

func main() {
	gin := route.Router()
	gin.Run(":8080")
}
