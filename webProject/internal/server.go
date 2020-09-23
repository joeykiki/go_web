package internal

import (
	"encoding/json"
	"log"
	"net"
	"reflect"
	"webProject/api"
)

type Server struct{
	Addr string
	Funcs map[string] reflect.Value
}

//新建服务端
func (server *Server) NewServer(){
	listen, err := net.Listen("tcp", server.Addr)
	if err != nil{
		log.Fatal("listen error", err)
		return
	}
	for{
		conn, err := listen.Accept()
		if err != nil{
			log.Fatal("accept error", err)
		}
		go server.HandleConn(conn)
	}
}

//处理连接
func (server *Server) HandleConn(conn net.Conn) {
	defer conn.Close()
	//读取request，传给Execute处理
	//获取返回体写回连接中
	transport := Transport{conn}
	for{
		buf, err := transport.Read()
		if err != nil{
			log.Println("read data error ", err)
		}
		log.Println("read request ", string(buf))
		var request api.Request
		err = json.Unmarshal(buf, &request)
		if err != nil{
			log.Println("json unmarshal error", err)
			return
		}

		response := server.Execute(request.FName, request.RequestBody)
		resp, err := json.Marshal(response)
		if err != nil{
			log.Fatal("json marshal error", err)
		}
		transport.Send(resp)
	}
}

func (server *Server) Execute(funcName string, reqBody interface{}) api.Response{
	f, _ := server.Funcs[funcName]
	inArgs := []reflect.Value{reflect.ValueOf(reqBody)}
	outArgs := f.Call(inArgs)

	respArgs := make([]interface{}, len(outArgs))
	for i := range outArgs{
		//调用interface返回实际值
		respArgs[i] = outArgs[i].Interface()
	}
	return api.Response{FName: funcName, ResponseBody: respArgs, Error: 0}

}

func (server *Server) Register(funcName string, fFunc interface{}){
	//判断是否已经注册了这个方法名
	if _, ok := server.Funcs[funcName]; ok{
		return
	}
	server.Funcs[funcName] = reflect.ValueOf(fFunc)
}