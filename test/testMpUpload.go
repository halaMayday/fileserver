package main

import (
	"bufio"
	"bytes"
	"fmt"
	jsonit "github.com/json-iterator/go"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func multipartUpload(filename, targetURL string, chunkSize int) error {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	bufferReader := bufio.NewReader(file)
	index := 0

	ch := make(chan int)
	buf := make([]byte, chunkSize)
	for {
		n, err := bufferReader.Read(buf)
		if n <= 0 {
			break
		}
		index++

		bufCopied := make([]byte, 5*1024*1024)
		copy(bufCopied, buf)

		go func(b []byte, curInx int) {
			fmt.Printf("upload_size:%d\n", len(b))

			resp, err := http.Post(
				targetURL+"&index="+strconv.Itoa(curInx),
				"multipart/form-data",
				bytes.NewReader(b))
			if err != nil {
				fmt.Println(err)
			}

			body, er := ioutil.ReadAll(resp.Body)
			fmt.Println("%+v %+v\n", string(body), er)
			resp.Body.Close()

			ch <- curInx
		}(bufCopied[:n], index)

		//遇到任何错误立即返回，并且忽略EOF错误
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err.Error())
			}
		}
	}

	for idx := 0; idx < index; idx++ {
		select {
		case res := <-ch:
			fmt.Println(res)
		}
	}
	return nil
}

func main() {
	username := "hufan"
	token := "xxxxx"
	filehash := ""

	//1.请求初始化分块上传接口
	resp, err := http.PostForm(
		"http://loaclhost:8080/file/mpupload/init",
		url.Values{
			"username": {username},
			"token":    {token},
			"filehash": {filehash},
			"filesize": {"132489256"}})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	//2.得到uploadID以及服务端指定的分块大小chunkSize
	uploadId := jsonit.Get(body, "data").Get("UploadID").ToString()
	chunkSize := jsonit.Get(body, "data").Get("chunkSize").ToInt()
	fmt.Printf("upload: %s   chunksize:%d\n", uploadId, chunkSize)

	//3.请求分块上传接口
	//需要替换为全路径。
	filename := "xxxxxx"
	tURL := "http://localhost:8080/file/mpupload/uppart?" +
		"username=admin&token=" + token + "&uploadid=" + uploadId
	multipartUpload(filename, tURL, chunkSize)

	//4.请求分块完成接口
	resp, err = http.PostForm(
		"http://localhost:8080/file/mpupload/complete",
		url.Values{
			"username": {username},
			"token":    {token},
			"filehash": {filehash},
			"filesize": {"132489256"},
			"filename": {"go1.10.3.linux-amd64.tar.gz"}, //需要替换文件名字
			"uploadid": {uploadId},
		})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	fmt.Printf("complete result: %s\n", string(body))
}
