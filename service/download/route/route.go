package route

import (
	"filestore-server/service/download/api"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//Router:路由表配置
func Router() *gin.Engine {

	router := gin.Default()
	//处理静态资源
	router.Static("/static/", "./static")

	//使用gin插件支持跨域
	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"}, // []string{"http://localhost:8080"},
		AllowMethods:  []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Range", "x-requested-with", "content-Type"},
		ExposeHeaders: []string{"Content-Length", "Accept-Ranges", "Content-Range", "Content-Disposition"},
	}))

	//文件下载相关接口
	router.GET("/file/download", api.DownloadHandler)
	router.POST("/file/downloadurl", api.DownloadURLHandler)

	return router
}
