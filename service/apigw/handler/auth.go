package handler

import (
	"filestore-server/common"
	"filestore-server/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

//todo：http请求拦截器
func IsTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}
	//TODO:判断token的时效性
	//TODO:从数据库tbl_user_token中查询username对应的token信息
	//TODO:对比两个TOKEN是否一致
	return true
}

//Authorize:http请求器
func Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")
		//验证登录token是否有效
		if len(username) < 3 || !IsTokenValid(token) {
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
