package handler

import (
	dblayer "filestore-server/db"
	"filestore-server/util"
	"fmt"
	"net/http"
	"time"
)

const (
	pwdSalt = "*#890"
)

//处理用户注册请求
func SignupInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/static/view/signup.html", http.StatusFound)
		return
	}
	r.ParseForm()

	username := r.Form.Get("username")
	passwd := r.Form.Get("password")

	//1.校验用户名和密码
	if len(username) < 3 || len(passwd) < 5 {
		w.Write([]byte("Invalid parameter"))
		return
	}

	//2.对密码进行加密

	encPassword := util.Sha1([]byte(passwd + pwdSalt))

	suc := dblayer.UserSingUp(username, encPassword)

	if suc {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("FAILED"))
	}

}

//登录接口
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/static/view/signup.html", http.StatusFound)
		return
	}

	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	encPassword := util.Sha1([]byte(password + pwdSalt))

	//1.校验用户名和密码
	pwdChecked := dblayer.UserSignin(username, encPassword)
	if !pwdChecked {
		w.Write([]byte("USERNAME OR PASSWORD FAILED"))
	}
	//2.生成访问凭证
	token := GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		w.Write([]byte("FAILED"))
		return
	}
	//3.登录成功后重定向到首页
	//w.Write([]byte("http://"+r.Host+"/static/view/home.html"))
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())

}

//UserInfoHandler:查询用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	//1.解析参数
	r.ParseForm()
	username := r.Form.Get("username")

	//2.查询用户信息
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	//3.组装并且相应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}

	w.Write(resp.JSONBytes())
}

//GenToken:生成token
func GenToken(username string) string {
	//40位字符：md5(username+timestamp+token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

//IsTokenValid:验证token有效性
func IsTokenValid(token string, username string) bool {
	if len(token) < 40 {
		return false
	}
	//判断token的时效性 token 有效期为半个小时
	tokenApplyTime := token[32:]
	//
	ts := fmt.Sprintf("%x", time.Now().Unix()-30*60)
	if tokenApplyTime < ts {
		fmt.Println("token is expire")
		return false
	}
	//查询username对应的token信息
	userToken := dblayer.GetUserToken(username)
	//对比两个token是否一致
	if token != userToken {
		fmt.Println("token is invalid 请重新登录")
		return false
	}
	return true
}
