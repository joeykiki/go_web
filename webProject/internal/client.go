package internal

import (
	"encoding/json"
	"fmt"
	//"google.golang.org/grpc/codes"
	"log"
	"net"
	"webProject/api"
	//"sync"
)

//var callLock sync.Mutex
//
//func withLock(f func()){
//	callLock.Lock()
//	defer callLock.Unlock()
//	f()
//}

const connections = 50

type Client struct{
	pool chan net.Conn
}

//客户端启动一个连接
func (client *Client) Dial(network, address string) error{
	client.pool = make(chan net.Conn, connections)
	for i := 0; i < connections; i++{
		conn, err := net.Dial(network, address)
		if err != nil{
			return err
		}
		client.pool <- conn
	}

	return nil
}

func (client *Client) getConn()(conn net.Conn, err error){
	select{
		case conn := <- client.pool:
			return conn, err
	}
}

func (client *Client) releaseConn(conn net.Conn)error{
	select{
		case client.pool <- conn:
			return nil
	}
}

//登录验证
func (client *Client) Login(request api.LoginRequest) *api.LoginResponse{
	response := client.CallRPC("Login", request)
	resp := new(api.LoginResponse)
	if response == nil{
		resp.Err = -1
	}else{
		resp.Err = int(response.([]interface{})[0].(map[string]interface{})["Err"].(float64))
	}
	return resp
}

//获取用户信息，将返回体封装成response
func (client *Client) GetInfo(request api.GetInfoRequest) *api.GetInfoResponse{
	response := client.CallRPC("GetInfo", request)
	resp := new(api.GetInfoResponse)
	if response == nil{
		resp.Err = -1
	}else{
		//unmarshal之后int变float64
		resp.Username = fmt.Sprintf("%v", response.([]interface{})[0].(map[string]interface{})["Username"])
		resp.Nickname = fmt.Sprintf("%v", response.([]interface{})[0].(map[string]interface{})["Nickname"])
		resp.Password = fmt.Sprintf("%v", response.([]interface{})[0].(map[string]interface{})["Password"])
		resp.Photo = fmt.Sprintf("%v", response.([]interface{})[0].(map[string]interface{})["Photo"])
		resp.Err = int(response.([]interface{})[0].(map[string]interface{})["Err"].(float64))
	}
	return resp
}

//调用鉴权服务
func (client *Client) TokenAuth(request api.AuthRequest) *api.AuthResponse{
	response := client.CallRPC("TokenAuth", request)
	resp := new(api.AuthResponse)
	if response == nil{
		resp.Err = -1
	}else{
		resp.Username = fmt.Sprintf("%v", response.([]interface{})[0].(map[string]interface{})["Username"])
		resp.Err = int(response.([]interface{})[0].(map[string]interface{})["Err"].(float64))
	}
	return resp
}

//修改数据
func (client *Client) Update(request api.UpdateRequest) *api.UpdateResponse{
	response := client.CallRPC("Update", request)
	resp := new(api.UpdateResponse)
	if response == nil{
		resp.Err = -1
	}else{
		resp.Err = int(response.([]interface{})[0].(map[string]interface{})["Err"].(float64))
	}
	return resp
}

//存储Token
func (client *Client) TokenSave(request api.TokenSaveRequest) *api.TokenSaveResponse{
	response := client.CallRPC("TokenSave", request)
	resp := new(api.TokenSaveResponse)
	if response == nil{
		resp.Err = -1
	}else{
		resp.Err = int(response.([]interface{})[0].(map[string]interface{})["Err"].(float64))
	}
	return resp
}

//解包
func (client *Client) CallRPC(funcName string, reqArgs interface{}) interface{}{
	request := api.Request{FName: funcName, RequestBody: reqArgs}
	response := client.Call(request)
	if response.Error != 0{
		return nil
	}

	//ResponseBody是map[string]interface{}类型
	return response.ResponseBody
}

//调用服务端方法
func (client *Client) Call(args api.Request) *api.Response{
	//沿着conn发送给服务端
	buf, err := json.Marshal(args)
	if err != nil{
		log.Fatal("client call error", err)
	}

	response := new(api.Response)

	//调用transport
	conn, _ := client.getConn()
	defer client.releaseConn(conn)
	transport := Transport{conn}
	transport.Send(buf)
	log.Println("send buf ", string(buf))
	data, err := transport.Read()

	if err != nil{
		log.Println("client read error", err)
		response.FName = args.FName
		response.ResponseBody = nil
		response.Error = -1
		return response
	}
	json.Unmarshal(data, response)

	return response
}
