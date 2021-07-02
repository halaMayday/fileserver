package handler

import (
	"filestore-server/common"
	"filestore-server/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

// HTTPInterceptor : http请求拦截器
func HTTPInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")
		//验证登录token是否有效
		if len(username) < 3 || !IsTokenValid(token, username) {
			// token校验失败则直接返回失败提示
			c.Abort()
			respMsg := util.NewRespMsg(
				int(common.StatusTokenInvalid),
				"token 无效",
				nil,
			)
			c.JSON(http.StatusOK, respMsg)
			return
		}
		c.Next()
	}
}
