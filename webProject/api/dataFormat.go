package api


const TCP_Addr = "localhost:8000"
const HTTP_Addr = "localhost:3000"
const REDIS_Addr = "localhost:6379"
const MYSQL_Source = "root:liujinru@/entry_task?charset=utf8"

const AUTH_Failed = 1

type Request struct{
	FName       string
	RequestBody interface{}
}

type Response struct{
	FName string
	ResponseBody interface{}
	Error int
}

type LoginRequest struct{
	Username, Password string
}
type LoginResponse struct{
	Err int
}

type GetInfoRequest struct{
	Username string
}
type GetInfoResponse struct{
	Username, Password, Nickname, Photo string
	Err int
}

type AuthRequest struct{
	Token string
}
type AuthResponse struct{
	Username string
	Err int
}

type UpdateRequest struct{
	Username, Nickname, Photo string
}
type UpdateResponse struct{
	Err int
}

type TokenSaveRequest struct{
	Username, Token string
}
type TokenSaveResponse struct{
	Err int
}
