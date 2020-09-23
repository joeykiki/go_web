package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"webProject/api"
	"webProject/internal"
)

var client *internal.Client

func main(){

	//启动客户端
	client = new(internal.Client)
	err := client.Dial("tcp", api.TCP_Addr)
	if err != nil{
		log.Println("connect tcp server error", err)
		return
	}

	http.Handle("/", http.FileServer(http.Dir("../web")))
	http.HandleFunc("/login", handlerLogin)
	http.HandleFunc("/update", handlerUpdate)
	http.HandleFunc("/auth", handlerTokenAuth)
	http.HandleFunc("/getinfo", handlerGetInfo)
	fmt.Println("running at port 3000")
	err = http.ListenAndServe(api.HTTP_Addr, nil)
	if err != nil{
		log.Fatal(err.Error())
	}
}

//处理前端传来的鉴权请求
func handlerTokenAuth(writer http.ResponseWriter, request *http.Request){
	token, err := request.Cookie("Token")
	var resp *api.AuthResponse
	if err != nil{
		resp = new(api.AuthResponse)
		resp.Err = -1
	}else{
		req := new(api.AuthRequest)
		req.Token = token.Value
		resp = client.TokenAuth(*req)
	}
	respByte := new(bytes.Buffer)
	json.NewEncoder(respByte).Encode(resp)
	writer.Write(respByte.Bytes())
}

//处理登录请求
func handlerLogin(writer http.ResponseWriter, request *http.Request){
	params := make(map[string]string)

	//解析json，并存入params map
	params["username"] = request.FormValue("username")
	params["password"] = request.FormValue("password")

	//定义请求体
	loginReq := new(api.LoginRequest)
	loginReq.Username = params["username"]
	loginReq.Password = params["password"]

	getInfoReq := new(api.GetInfoRequest)
	getInfoReq.Username = params["username"]

	//接收返回体
	respByte := new(bytes.Buffer)
	var getInfoResp *api.GetInfoResponse

	loginResp := client.Login(*loginReq)
	if loginResp.Err != 0{
		getInfoResp = new(api.GetInfoResponse)
		getInfoResp.Err = -1
	}else {
		getInfoResp = client.GetInfo(*getInfoReq)
		//username+time 转 base64
		t := time.Now()
		token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s%d", params["username"], t.Unix())))
		//异步存Token
		go func(username, token string){
			saveReq := new(api.TokenSaveRequest)
			saveReq.Username = username
			saveReq.Token = token
			saveResp := client.TokenSave(*saveReq)
			if saveResp.Err != 0{
				log.Println("save token error")
			}
		}(params["username"], token)

		//往ResponseWriter写Cookie
		cookie := http.Cookie{Name: "Token", Value: token}
		http.SetCookie(writer, &cookie)
	}
	json.NewEncoder(respByte).Encode(getInfoResp)
	writer.Write(respByte.Bytes())
}

//处理修改请求
func handlerUpdate(writer http.ResponseWriter, request *http.Request){
	//调用鉴权服务
	token, err := request.Cookie("Token")
	var respAuth *api.AuthResponse
	if err != nil{
		respAuth = new(api.AuthResponse)
		respAuth.Err = -1
	}else{
		req := new(api.AuthRequest)
		req.Token = token.Value
		respAuth = client.TokenAuth(*req)
	}
	var username string
	if respAuth.Err == -1{
		resp := new(api.UpdateResponse)
		resp.Err = api.AUTH_Failed
		respByte := new(bytes.Buffer)
		json.NewEncoder(respByte).Encode(resp)
		writer.Write(respByte.Bytes())
		return
	}else{
		username = respAuth.Username
	}


	params := make(map[string]string)

	//接收图片文件
	err = request.ParseMultipartForm(100000)
	if err != nil{
		log.Println("parse multipartform error ", err)
		return
	}
	m := request.MultipartForm
	if m.File["photo"] == nil{
		params["photo"] = ""
	}else{
		file := m.File["photo"][0]
		fopen, err := file.Open()
		defer fopen.Close()
		if err != nil{
			log.Println("file open error ", err)
			return
		}
		destPath, err := os.Create("../web/uimages/" + file.Filename)
		defer destPath.Close()
		if err != nil{
			log.Println("create image path error ", err)
			return
		}
		if _, err = io.Copy(destPath, fopen); err != nil{
			log.Println("copy image error ", err)
			return
		}
		params["photo"] = "uimages/" + file.Filename
	}

	//解析json，并存入params map
	params["username"] = username
	params["nickname"] = request.FormValue("nickname")

	//定义请求体
	req := new(api.UpdateRequest)
	req.Username = params["username"]
	req.Nickname = params["nickname"]
	req.Photo = params["photo"]

	//接收返回体
	resp := client.Update(*req)
	respByte := new(bytes.Buffer)
	json.NewEncoder(respByte).Encode(resp)
	writer.Write(respByte.Bytes())
}

//处理获取信息请求
func handlerGetInfo(writer http.ResponseWriter, request *http.Request){
	//调用鉴权服务
	token, err := request.Cookie("Token")
	var respAuth *api.AuthResponse
	if err != nil{
		respAuth = new(api.AuthResponse)
		respAuth.Err = -1
	}else{
		req := new(api.AuthRequest)
		req.Token = token.Value
		respAuth = client.TokenAuth(*req)
	}
	var username string
	if respAuth.Err == -1{
		resp := new(api.UpdateResponse)
		resp.Err = api.AUTH_Failed
		respByte := new(bytes.Buffer)
		json.NewEncoder(respByte).Encode(resp)
		writer.Write(respByte.Bytes())
		return
	}else{
		username = respAuth.Username
	}

	params := make(map[string]string)
	params["username"] = username

	req := new(api.GetInfoRequest)
	req.Username = params["username"]
	resp := client.GetInfo(*req)
	respByte := new(bytes.Buffer)
	json.NewEncoder(respByte).Encode(resp)
	writer.Write(respByte.Bytes())
}
