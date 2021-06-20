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
	http.HandleFunc("/file/upload", handler.HTTPInterceptor(handler.UploadHandler))
	http.HandleFunc("/file/fastupload", handler.HTTPInterceptor(handler.TryFastUploadHandler))

	http.HandleFunc("/file/upload/suc", handler.HTTPInterceptor(handler.UploadSucHandler))
	http.HandleFunc("/file/query", handler.HTTPInterceptor(handler.QueryFilehandler))
	http.HandleFunc("/file/download", handler.HTTPInterceptor(handler.DownloadHandler))
	http.HandleFunc("/file/update", handler.HTTPInterceptor(handler.FileMetaUpdataHandle))
	http.HandleFunc("/file/delete", handler.HTTPInterceptor(handler.FileDeletaHandle))
	http.HandleFunc("/file/downloadurl", handler.HTTPInterceptor(handler.DownloadHandler))

	// 用户相关接口
	//http.HandleFunc("/", handler.SignInHandler)
	http.HandleFunc("/user/signup", handler.HTTPInterceptor(handler.SignupInHandler))
	http.HandleFunc("/user/signin", handler.HTTPInterceptor(handler.SignInHandler))
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler))

	//分块上传接口
	http.HandleFunc("/file/mpupload/init",
		handler.HTTPInterceptor(handler.InitalMultipartUploadHandler))
	http.HandleFunc("/file/mpupload/uppart",
		handler.HTTPInterceptor(handler.UploadPartHandler))
	http.HandleFunc("/file/mpupload/finshed",
		handler.HTTPInterceptor(handler.CompleteUploadHandler))
	http.HandleFunc("/file/mpupload/cancel",
		handler.HTTPInterceptor(handler.CancelUploadPartHandler))
	http.HandleFunc("/file/mpupload/status",
		handler.HTTPInterceptor(handler.MultipartUploadStatusHanlder))

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Printf("Failed to start server,err:%s", err.Error())
	}
}
