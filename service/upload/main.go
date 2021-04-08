package main

import (
	"filestore-server/handler"
	"fmt"
	"net/http"
)

func main() {
	// 静态资源处理
	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(assets.AssetFS())))
	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// 文件存取接口
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/meta", handler.GetFileMetahandler)
	//http.HandleFunc("/file/query", handler.FileQueryHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/update", handler.FileMetaUpdataHandle)
	http.HandleFunc("/file/delete", handler.FileDeletaHandle)

	// 用户相关接口
	//http.HandleFunc("/", handler.SignInHandler)
	http.HandleFunc("/user/signup", handler.UserSignup)
	//http.HandleFunc("/user/signin", handler.SignInHandler)
	//http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Failed to start server,err:%s", err.Error())
	}
}
