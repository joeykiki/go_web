package internal

import(
	"database/sql"
	"fmt"
	"log"
	"github.com/garyburd/redigo/redis"
	"time"
	"webProject/api"

	_"github.com/go-sql-driver/mysql"
)

type DBConnection struct{
	DB *sql.DB
}

type User struct{
	Username string
	Nickname string
	Password string
	Photo string
}

func (dbConnection *DBConnection) NewConnection() error{
	db, err:= sql.Open("mysql", api.MYSQL_Source)
	if err != nil{
		log.Println("db conn error", err)
	}
	dbConnection.DB = db
	return err
}

func (dbConnection *DBConnection) Query(username string, password string) int{
	rows, err := dbConnection.DB.Query("SELECT * FROM user where username=? and password=?", username, password)
	if err != nil{
		log.Println("db query by username error", err)
		return -1
	}
	if rows.Next() == false{
		log.Println("db query by username return nil ")
		return -1
	}
	return 0
}

func (dbConnection *DBConnection) GetUserInfo(username string) *User{
	rows, err := dbConnection.DB.Query("SELECT * FROM user where username=?", username)
	if err != nil{
		log.Println("db query by username error", err)
		return nil
	}
	user := new(User)
	for rows.Next(){
		err = rows.Scan(&user.Username, &user.Nickname, &user.Password, &user.Photo)
		if err != nil{
			log.Println("rows scan error", err)
			return nil
		}
	}
	return user
}

func (dbConnection *DBConnection) Update(username, nickname, photo string) int{
	if photo == ""{
		_, err := dbConnection.DB.Exec("UPDATE user SET nickname=? WHERE username=?", nickname, username)
		if err != nil{
			log.Println("db update error", err)
			return -1
		}
	}else{
		_, err := dbConnection.DB.Exec("UPDATE user SET nickname=?, photo=? WHERE username=?", nickname, photo, username)
		if err != nil{
			log.Println("db update error", err)
			return -1
		}
	}
	return 0
}

var redisClient *redis.Pool

type RedisConnection struct{
	//Conn redis.Conn
	redisPool redis.Pool
}

func (redisConnection *RedisConnection) NewRedisConn(){
	redisConnection.redisPool = redis.Pool{
		MaxIdle: 50,
		MaxActive: 50,
		IdleTimeout: 200 * time.Second,
		Wait: true,
		Dial: func()(redis.Conn, error){
			con, err := redis.Dial("tcp", api.REDIS_Addr)
			if err != nil{
				return nil, err
			}
			return con, nil
		},
	}
}

func (redisConnection *RedisConnection) getConn() (conn redis.Conn, err error){
	return redisConnection.redisPool.Get(), nil
}

func (redisConnection *RedisConnection) releaseConn(conn redis.Conn){
	conn.Close()
}

func (redisConnection *RedisConnection) QueryToken(token string) string{
	conn, _ := redisConnection.getConn()
	defer redisConnection.releaseConn(conn)
	username, err := redis.String(conn.Do("GET", token))
	if err != nil{
		log.Println("redis query error ", err)
		return username
	}
	return username
}

func (redisConnection *RedisConnection) SaveToken(username, token string) int{
	conn, _ := redisConnection.getConn()
	defer redisConnection.releaseConn(conn)
	_, err := conn.Do("SET", token, username)
	if err != nil{
		log.Println("redis save token error ", err)
		return -1
	}

	//设置过期时间，15min
	_, err = conn.Do("EXPIRE", token, 900)
	if err != nil{
		log.Println("set expire time error", err)
	}
	return 0
}

func (redisConnection *RedisConnection) GetPassword(username string) string{
	conn, _ := redisConnection.getConn()
	defer redisConnection.releaseConn(conn)
	result, err := redis.String(conn.Do("HGET", username, "password"))
	if err != nil{
		log.Println("redis get password error ", err)
		return ""
	}
	return result
}

func (redisConnection *RedisConnection) SetPassword(username, password string) int{
	conn, _ := redisConnection.getConn()
	defer redisConnection.releaseConn(conn)
	_, err := conn.Do("HSET", username, "password", password)
	if err != nil{
		log.Println("redis set password error ", err)
		return -1
	}
	return 0
}

func (redisConnection *RedisConnection) GetUserInfo(username string) *User{
	conn, _ := redisConnection.getConn()
	defer redisConnection.releaseConn(conn)
	exist, err := redis.Bool(conn.Do("EXISTS", "info_"+username))
	if exist == false{
		return nil
	}
	result, err := redis.Strings(conn.Do("HMGET", "info_"+username, "nickname", "photo"))
	if err != nil{
		log.Println("redis get user info by username error", err)
		return nil
	}
	fmt.Println("redis get user info ", result)
	user := new(User)
	user.Username = username
	user.Nickname = result[0]
	user.Photo = result[1]

	return user
}

func (redisConnection *RedisConnection) SetUserInfo(username, nickname, photo string) int{
	conn, _ := redisConnection.getConn()
	defer redisConnection.releaseConn(conn)
	if photo == ""{
		_, err := conn.Do("HMSET", "info_"+username, "nickname", nickname)
		if err != nil{
			log.Println("redis set user info error ", err)
			return -1
		}
	}else{
		_, err := conn.Do("HMSET", "info_"+username, "nickname", nickname, "photo", photo)
		if err != nil{
			log.Println("redis set user info error ", err)
			return -1
		}
	}
	return 0
}

func (redisConnection *RedisConnection) Update(username, nickname, photo string) int{
	conn, _ := redisConnection.getConn()
	defer redisConnection.releaseConn(conn)
	_, err := conn.Do("HMSET", "info_" + username, "nickname", nickname, "photo", photo)
	if err != nil{
		log.Println("redis update err")
		return -1
	}else{
		return 0
	}
}

func (redisConnection *RedisConnection) DelUserInfo(username string) int{
	conn, _ := redisConnection.getConn()
	defer redisConnection.releaseConn(conn)
	_, err := conn.Do("DEL", "info_" + username)
	if err != nil{
		log.Println("redis delete err ", err)
		return -1
	}
	return 0
}
