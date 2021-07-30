package route

import (
	"filestore-server/service/upload/api"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	router := gin.Default()

	router.Static("/static", "./static")

	//使用gin处理跨域请求
	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"}, // []string{"http://localhost:8080"},
		AllowMethods:  []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Range", "x-requested-with", "content-Type"},
		ExposeHeaders: []string{"Content-Length", "Accept-Ranges", "Content-Range", "Content-Disposition"},
		// AllowCredentials: true,

	}))

	//文件存取接口

	// 文件上传相关接口
	router.POST("/file/upload", api.DoUploadHandler)
	// 秒传接口
	router.POST("/file/fastupload", api.TryFastUploadHandler)

	// 分块上传接口
	router.POST("/file/mpupload/init", api.InitalMultipartUploadHandler)
	router.POST("/file/mpupload/uppart", api.UploadPartHandler)
	router.POST("/file/mpupload/complete", api.CompleteUploadHandler)

	return router
}
