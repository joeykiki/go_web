package main

import (
	"fmt"
	"log"
	"reflect"
	"webProject/api"
	"webProject/internal"
	"crypto/md5"
)

//登录功能
func Login(request map[string] interface{}) *api.LoginResponse{
	log.Println("login happens")
	var username = request["Username"].(string)
	var password = fmt.Sprintf("%x", md5.Sum([]byte(request["Password"].(string))))

	retCode := loginHelper(username, password)
	resp := new(api.LoginResponse)
	resp.Err = retCode
	return resp
}
func loginHelper(username, password string) int{
	pwd := RedisConn.GetPassword(username)
	if pwd == ""{
		retCode := MysqlConn.Query(username, password)
		if retCode == 0{
			//同步到redis
			RedisConn.SetPassword(username, password)
		}
		return retCode
	}else{
		if password == pwd{
			return 0
		}else{
			return -1
		}
	}
}

//获取用户信息
func GetInfo(request map[string]interface{}) *api.GetInfoResponse{
	log.Println("get info happens")
	var username = request["Username"].(string)

	user := new(internal.User)
	user = GetInfoHelper(username)
	resp := new(api.GetInfoResponse)
	if user != nil{
		resp.Username = user.Username
		resp.Nickname = user.Nickname
		resp.Password = user.Password
		resp.Photo = user.Photo
		resp.Err = 0
	}else{
		resp.Err = -1
	}
	return resp
}
func GetInfoHelper(username string) *internal.User{
	user := RedisConn.GetUserInfo(username)
	if user == nil{
		user = MysqlConn.GetUserInfo(username)
		//同步到redis
		RedisConn.SetUserInfo(username, user.Nickname, user.Photo)
	}
	return user
}

//鉴权功能
func TokenAuth(request map[string]interface{}) *api.AuthResponse{
	log.Println("token auth happens")
	var token = request["Token"].(string)

	username := RedisConn.QueryToken(token)
	resp := new(api.AuthResponse)
	if username != ""{
		resp.Username = username
		resp.Err = 0
	}else{
		resp.Err = -1
	}
	return resp
}

//修改功能
func Update(request map[string]interface{}) *api.UpdateResponse{
	log.Println("update happens")
	var username = request["Username"].(string)
	var nickname = request["Nickname"].(string)
	var photo = request["Photo"].(string)

	retCode := UpdateHelper(username, nickname, photo)
	resp := new(api.UpdateResponse)
	resp.Err = retCode
	return resp
}
func UpdateHelper(username, nickname, photo string) int{
	retCode := MysqlConn.Update(username, nickname, photo)
	RedisConn.DelUserInfo(username)
	return retCode
}

//存储Token
func TokenSave(request map[string]interface{}) *api.TokenSaveResponse{
	log.Println("token save happens")
	var username = request["Username"].(string)
	var token = request["Token"].(string)

	retCode := RedisConn.SaveToken(username, token)
	resp := new(api.TokenSaveResponse)
	resp.Err = retCode
	return resp
}

var MysqlConn internal.DBConnection
var RedisConn internal.RedisConnection

func main(){

	//启动服务器，并监听
	server := internal.Server{Addr:api.TCP_Addr, Funcs: nil}
	server.Funcs = make(map[string]reflect.Value)

	//注册方法
	server.Register("Login", Login)
	server.Register("TokenAuth", TokenAuth)
	server.Register("Update", Update)
	server.Register("GetInfo", GetInfo)
	server.Register("TokenSave", TokenSave)

	//创建数据库连接
	err := MysqlConn.NewConnection()
	defer MysqlConn.DB.Close()
	if err != nil{
		return
	}
	RedisConn.NewRedisConn()

	server.NewServer()
}
