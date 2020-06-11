package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"

	"io"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gorilla/websocket"
	"github.com/lib/pq"
	"github.com/urfave/negroni"
)

const (
	//SecretKey tokenKey
	SecretKey = "sxbvsf56"
	host      = "localhost"
	port      = 4000
	user      = "root"
	password  = "passwd"
	dbname    = "chat"
)

//Users 表结构
type Users struct {
	email       string
	username    string
	phonenumber string
	password    string
	isAdmin     bool
}

//Client 客户端
type Client struct {
	//用户id
	username string
	//连接的socket
	socket *websocket.Conn
	//发送的消息
	send chan []byte
}

//Message 会把Message格式化成json
type Message struct {
	//消息struct
	Sender   string `json:"sender"`   //发送者
	Receiver string `json:"receiver"` //接收者
	Content  string `json:"content"`  //内容
	File     string `json:"file"`
	Filesrc  string `json:"filesrc"`
	Date     string `json:"date"`
}

//创建客户端管理者
var manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

//ClientManager 客户端管理
type ClientManager struct {
	//客户端 map 储存并管理所有的长连接client，在线的为true，不在的为false
	clients map[*Client]bool
	//web端发送来的的message我们用broadcast来接收，并最后分发给所有的client
	broadcast chan []byte
	//新创建的长连接client
	register chan *Client
	//新注销的长连接client
	unregister chan *Client
}

//addUserSQL 添加用户
func addUserSQL(db *sql.DB, info Users) error {
	sqlStatement := `INSERT INTO users (username, email, phonenumber, password, "isAdmin")
	VALUES ($1, $2, $3, $4, $5)RETURNING uid`
	_, err := db.Exec(sqlStatement, info.username, info.email, info.phonenumber, info.password, info.isAdmin)

	errif(err)
	return nil
}

//UpdateUserSQL 更新用户
func UpdateUserSQL(db *sql.DB, info Users) error {
	sqlStatement := `UPDATE users SET email = $2, phonenumber = $3, password = $4, "isAdmin" = $5 WHERE username = $1;`
	_, err := db.Exec(sqlStatement, info.username, info.email, info.phonenumber, info.password, info.isAdmin)

	errif(err)
	return nil
}

//DeleteUserSQL 删除用户
func DeleteUserSQL(db *sql.DB, Username string) error {
	sqlStatement := `DELETE FROM users WHERE username = $1;`
	_, err := db.Exec(sqlStatement, Username)

	errif(err)
	return nil
}

//QueryUserSQL 查询用户
func QueryUserSQL(db *sql.DB, authUser Users) Users {
	sqlStatement := `SELECT email, phonenumber, password, "isAdmin" FROM users WHERE username=$1;`
	row := db.QueryRow(sqlStatement, authUser.username)
	QueryRes := authUser
	err := row.Scan(&QueryRes.email, &QueryRes.phonenumber, &QueryRes.password, &QueryRes.isAdmin)
	if err == nil {
		return QueryRes
	}

	QueryRes = authUser
	sqlStatement = `SELECT username, phonenumber, password, "isAdmin" FROM users WHERE email=$1;`
	row = db.QueryRow(sqlStatement, authUser.email)
	err = row.Scan(&QueryRes.username, &QueryRes.phonenumber, &QueryRes.password, &QueryRes.isAdmin)
	if err == nil {
		return QueryRes
	}

	QueryRes = authUser
	sqlStatement = `SELECT email, username, password, "isAdmin" FROM users WHERE phonenumber=$1;`
	row = db.QueryRow(sqlStatement, authUser.phonenumber)
	err = row.Scan(&QueryRes.email, &QueryRes.username, &QueryRes.password, &QueryRes.isAdmin)
	if err == nil {
		return QueryRes
	}

	QueryRes.username = ""
	return QueryRes
}

//checkUser 检查是否已存在用户
func checkUser(db *sql.DB, info Users) string {
	if info.username != "" {
		sqlStatement := `SELECT "isAdmin" FROM users WHERE username=$1;`
		row := db.QueryRow(sqlStatement, info.username)
		err := row.Scan(&info.isAdmin)
		println(err)
		if err == nil {
			return "existed"
		}
	}

	if info.phonenumber != "" {
		sqlStatement := `SELECT "isAdmin" FROM users WHERE phonenumber=$1;`
		row := db.QueryRow(sqlStatement, info.phonenumber)
		err := row.Scan(&info.isAdmin)
		println(err)
		if err == nil {
			return "existed"
		}
	}

	if info.email != "" {
		sqlStatement := `SELECT "isAdmin" FROM users WHERE email=$1;`
		row := db.QueryRow(sqlStatement, info.email)
		err := row.Scan(&info.isAdmin)
		println(err)
		if err == nil {
			return "existed"
		}
	}
	return "does not exist"
}

//reJSON 返回JSON
func reJSON(writer http.ResponseWriter, data *map[string]interface{}) {
	js, err := json.Marshal(data)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(js)
}

func errif(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

func authAccount(writer http.ResponseWriter, request *http.Request) {
	authUser := Users{
		username:    string(request.FormValue("username")),
		email:       string(request.FormValue("username")),
		phonenumber: string(request.FormValue("username")),
		password:    string(request.FormValue("password")),
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	errif(err)
	defer db.Close()

	temp := QueryUserSQL(db, authUser)

	if temp.username == "" {
		res := map[string]interface{}{"Status": "does not exist"}
		reJSON(writer, &res)

	} else if authUser.password == temp.password {
		token := jwt.New(jwt.SigningMethodHS256)
		claims := make(jwt.MapClaims)
		claims["exp"] = time.Now().Add(time.Hour * time.Duration(1)).Unix()
		claims["nbf"] = time.Now().Unix()
		claims["username"] = authUser.username
		claims["isAdmin"] = temp.isAdmin
		token.Claims = claims
		tokenString, _ := token.SignedString([]byte(SecretKey))

		res := map[string]interface{}{"Status": "Verified", "token": tokenString}
		reJSON(writer, &res)

	} else {
		res := map[string]interface{}{"Status": "Verification failed"}
		reJSON(writer, &res)
	}
}

func addUser(writer http.ResponseWriter, request *http.Request) {
	addTemp := Users{
		email:       string(request.FormValue("email")),
		username:    string(request.FormValue("username")),
		password:    string(request.FormValue("password")),
		phonenumber: string(request.FormValue("phonenumber")),
		isAdmin:     false,
	}
	addTemp.isAdmin, _ = strconv.ParseBool(request.FormValue("isAdmin"))

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	errif(err)
	defer db.Close()
	temp := QueryUserSQL(db, addTemp)

	if addTemp.username != temp.username {
		token := jwt.New(jwt.SigningMethodHS256)
		claims := make(jwt.MapClaims)
		claims["exp"] = time.Now().Add(time.Hour * time.Duration(1)).Unix()
		claims["iat"] = time.Now().Unix()
		claims["username"] = addTemp.username
		claims["isAdmin"] = temp.isAdmin
		token.Claims = claims
		tokenString, _ := token.SignedString([]byte(SecretKey))

		err = addUserSQL(db, addTemp)
		errif(err)
		res := map[string]interface{}{"Status": "Succeeded", "token": tokenString}
		reJSON(writer, &res)

	} else {
		res := map[string]interface{}{"Status": "Failed"}
		reJSON(writer, &res)
	}
}

func updateUser(writer http.ResponseWriter, request *http.Request) {
	updateTemp := Users{
		username:    string(request.FormValue("username")),
		email:       string(request.FormValue("email")),
		phonenumber: string(request.FormValue("phonenumber")),
		password:    string(request.FormValue("password")),
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	errif(err)
	defer db.Close()
	temp := QueryUserSQL(db, updateTemp)

	if temp.username != "" {
		updateTemp.isAdmin = temp.isAdmin
		if updateTemp.password == "" {
			updateTemp.password = temp.password
		}
		err := UpdateUserSQL(db, updateTemp)
		errif(err)
		res := map[string]interface{}{"Status": "Succeeded"}
		reJSON(writer, &res)

	} else {
		res := map[string]interface{}{"Status": "Failed"}
		reJSON(writer, &res)
	}
}

func deleteUser(writer http.ResponseWriter, request *http.Request) {
	deleteTemp := Users{
		username: string(request.FormValue("username")),
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	errif(err)
	defer db.Close()
	temp := QueryUserSQL(db, deleteTemp)

	if temp.username == deleteTemp.username {
		err = DeleteUserSQL(db, deleteTemp.username)
		errif(err)
		res := map[string]interface{}{"Status": "Succeeded"}
		reJSON(writer, &res)

	} else {
		res := map[string]interface{}{"Status": "Failed"}
		reJSON(writer, &res)
	}
}

//UserList 用户列表
func UserList(writer http.ResponseWriter, request *http.Request) {
	if ValidateTokenMiddleware(writer, request) == "OK" {
		psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
		db, err := sql.Open("postgres", psqlInfo)
		errif(err)
		defer db.Close()

		sqlStatement := `select username, email, phonenumber, "isAdmin" from users;`
		rows, _ := db.Query(sqlStatement)
		defer rows.Close()

		Userlist := []map[string]interface{}{}
		var temp Users
		for rows.Next() {
			_ = rows.Scan(&temp.username, &temp.email, &temp.phonenumber, &temp.isAdmin)
			t1 := map[string]interface{}{"username": temp.username, "email": temp.email, "phonenumber": temp.phonenumber, "isAdmin": temp.isAdmin}
			Userlist = append(Userlist, t1)
		}
		js, err := json.Marshal(Userlist)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(js)
	} else {
		writer.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(writer, "Unauthorized access to this resource")
	}
}

func chatList(writer http.ResponseWriter, request *http.Request) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	errif(err)
	defer db.Close()

	sqlStatement := `select username from users`
	rows, _ := db.Query(sqlStatement)
	defer rows.Close()

	Userlist := []map[string]interface{}{}
	var temp string
	for rows.Next() {
		_ = rows.Scan(&temp)
		t1 := map[string]interface{}{"username": temp}
		Userlist = append(Userlist, t1)
	}
	js, err := json.Marshal(Userlist)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(js)
}

func checkinfo(writer http.ResponseWriter, request *http.Request) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	errif(err)
	defer db.Close()

	var info Users
	queryParam := request.URL.Query()
	info.username = string(queryParam.Get("username"))
	info.email = string(queryParam.Get("email"))
	info.phonenumber = string(queryParam.Get("phonenumber"))
	res := checkUser(db, info)

	if res == "does not exist" {
		res := map[string]interface{}{"Status": "does not exist"}
		reJSON(writer, &res)
	} else {
		res := map[string]interface{}{"Status": "existed"}
		reJSON(writer, &res)
	}
}

//ValidateTokenMiddleware 验证token中间件
func ValidateTokenMiddleware(w http.ResponseWriter, r *http.Request) string {

	token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SecretKey), nil
		})

	if err == nil && token.Valid {
		return "OK"
	}
	return "dd"
}

//OPTIONSrequest 二次请求判断
func OPTIONSrequest(writer http.ResponseWriter, request *http.Request, next http.HandlerFunc) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	if request.Method == "POST" {
		next(writer, request)
	} else if request.Method == "GET" {
		next(writer, request)
	} else {
		writer.Header().Add("Access-Control-Allow-Headers", "authorization,content-type")
		writer.Header().Set("content-type", "application/json")
		writer.Header().Add("Access-Control-Allow-Methods", "PUT,GET,POST,DELETE,OPTIONS")
	}
}

func reCAPTCHA(writer http.ResponseWriter, r *http.Request) {
	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify",
		url.Values{"secret": {"6Ld6Y3sUAAAAAP2CzuEzQOO5D2kIptaz9YdF_BrI"}, "response": {r.FormValue("captchaResponse")}})
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(body)
}

func main() {
	http.HandleFunc("/chatlist", chatList)
	http.Handle("/auth", negroni.New(
		negroni.HandlerFunc(OPTIONSrequest),
		negroni.Wrap(http.HandlerFunc(authAccount)),
	))
	http.Handle("/userlist", negroni.New(
		negroni.HandlerFunc(OPTIONSrequest),
		negroni.Wrap(http.HandlerFunc(UserList)),
	))
	http.Handle("/addUser", negroni.New(
		negroni.HandlerFunc(OPTIONSrequest),
		negroni.Wrap(http.HandlerFunc(addUser)),
	))
	http.Handle("/deleteUser", negroni.New(
		negroni.HandlerFunc(OPTIONSrequest),
		negroni.Wrap(http.HandlerFunc(deleteUser)),
	))
	http.Handle("/updateUser", negroni.New(
		negroni.HandlerFunc(OPTIONSrequest),
		negroni.Wrap(http.HandlerFunc(updateUser)),
	))
	http.Handle("/checkinfo", negroni.New(
		negroni.HandlerFunc(OPTIONSrequest),
		negroni.Wrap(http.HandlerFunc(checkinfo)),
	))
	http.HandleFunc("/reCAPTCHA", reCAPTCHA)
	go manager.start()
	//注册默认路由为 /ws ，并使用wsHandler这个方法
	http.HandleFunc("/wss", wsHandler)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/chathst", chatSQL)
	http.ListenAndServe(":3000", nil)

}

func (manager *ClientManager) start() {
	for {
		select {
		//如果有新的连接接入,就通过channel把连接传递给conn
		case conn := <-manager.register:
			//把客户端的连接设置为true
			manager.clients[conn] = true
		case conn := <-manager.unregister:
			//判断连接的状态，如果是true,就关闭send，删除连接client的值
			if _, ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
			}
			//广播
		case message := <-manager.broadcast:
			var temp Message
			json.Unmarshal(message, &temp)
			chattoSQL(temp)
			//遍历已经连接的客户端，把消息发送给指定客户端
			for conn := range manager.clients {
				if conn.username == temp.Receiver {
					select {
					case conn.send <- message:
					default:
						close(conn.send)
						delete(manager.clients, conn)
					}
				}
			}
		}
	}
}

//定义客户端管理的send方法
func (manager *ClientManager) send(message []byte, ignore *Client) {
	for conn := range manager.clients {
		//不给屏蔽的连接发送消息
		if conn != ignore {
			conn.send <- message
		}
	}
}

//定义客户端结构体的read方法
func (c *Client) read() {
	defer func() {
		manager.unregister <- c
		c.socket.Close()
	}()

	for {
		//读取消息
		_, message, err := c.socket.ReadMessage()
		//如果有错误信息，就注销这个连接然后关闭
		if err != nil {
			manager.unregister <- c
			c.socket.Close()
			break
		}
		//如果没有错误信息就把信息放入broadcast`
		manager.broadcast <- message
	}
}

func (c *Client) write() {
	defer func() {
		c.socket.Close()
	}()

	for {
		select {
		//从send里读消息
		case message, ok := <-c.send:
			//如果没有消息
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			//有消息就写入，发送给web端
			c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func wsHandler(writer http.ResponseWriter, request *http.Request) {
	//将http协议升级成websocket协议
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(writer, request, nil)
	if err != nil {
		http.NotFound(writer, request)
		return
	}

	client := &Client{
		username: request.FormValue("username"),
		socket:   conn,
		send:     make(chan []byte),
	}
	//注册一个新的链接
	manager.register <- client

	//启动协程收web端传过来的消息
	go client.read()
	//启动协程把消息返回给web端
	go client.write()
}

func chattoSQL(message Message) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	errif(err)
	defer db.Close()

	sqlStatement := `SELECT content FROM chat WHERE (users[1]=$1 AND users[2]=$2) OR (users[1]=$2 AND users[2]=$1);`
	rows, err := db.Query(sqlStatement, message.Sender, message.Receiver)
	if rows.Next() {
		sqlStatement := `UPDATE chat SET sender = array_append(sender, $1),
		receiver = array_append(receiver, $2),
		content = array_append(content, $3),
		file = array_append(file, $4),
		filesrc = array_append(filesrc, $5),
		date = array_append(date, $6)
		WHERE (users[1]=$1 AND users[2]=$2) OR (users[1]=$2 AND users[2]=$1);`
		_, err := db.Exec(sqlStatement, message.Sender, message.Receiver, message.Content, message.File, message.Filesrc, message.Date)

		errif(err)
	} else {
		temp := Message{
			Sender:   "{" + message.Sender + "}",
			Receiver: "{" + message.Receiver + "}",
			Content:  "{" + message.Content + "}",
			File:     "{" + message.File + "}",
			Filesrc:  "{" + message.Filesrc + "}",
			Date:     "{" + message.Date + "}",
		}
		ins := "{" + message.Sender + "," + message.Receiver + "}"
		sqlStatement := `INSERT INTO chat (users, sender, receiver, content, file, filesrc, date)
				VALUES ($1,$2,$3,$4,$5,$6,$7);`
		_, err := db.Exec(sqlStatement, ins, temp.Sender, temp.Receiver, temp.Content, temp.File, temp.Filesrc, temp.Date)

		errif(err)
	}
}

func upload(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	request.ParseMultipartForm(32 << 20)
	file, handler, err := request.FormFile("image")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	file1 := file
	imgBytes := make([]byte, handler.Size)
	bufr := bufio.NewReader(file1)
	_, err = bufr.Read(imgBytes)
	if err != nil {
		fmt.Println(err)
		res := map[string]interface{}{"Status": "failed"}
		reJSON(writer, &res)
		return
	}

	MD5 := md5.New()
	MD5.Write(imgBytes)
	md5str := hex.EncodeToString(MD5.Sum(nil))
	if md5str == "" {
		fmt.Println(fmt.Errorf("图片的md5获取失败"))
		return
	}
	filename := md5str + ".png"
	filePath := "../client/images/" + filename

	if !Exists(filePath) {
		f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, bytes.NewReader(imgBytes))
	}

	res := map[string]interface{}{"Status": "Succeeded", "url": "/images/" + filename}
	reJSON(writer, &res)
}

//Exists 检查是否有相同的图片
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

type chatsql struct {
	users    [2]string
	sender   []string
	receiver []string
	content  []string
	file     []string
	filesrc  []string
	date     []string
}

func chatSQL(writer http.ResponseWriter, request *http.Request) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	errif(err)
	defer db.Close()

	var chathst chatsql
	chathst.users[0] = request.URL.Query().Get("user1")
	chathst.users[1] = request.URL.Query().Get("user2")
	sqlStatement := `SELECT sender, receiver, content, file, filesrc, date FROM chat WHERE (users[1]=$1 AND users[2]=$2) OR (users[1]=$2 AND users[2]=$1);`
	row := db.QueryRow(sqlStatement, chathst.users[0], chathst.users[1])
	err = row.Scan(pq.Array(&chathst.sender), pq.Array(&chathst.receiver), pq.Array(&chathst.content), pq.Array(&chathst.file), pq.Array(&chathst.filesrc), pq.Array(&chathst.date))
	if err != nil {
		res := map[string]interface{}{"content": "null"}
		reJSON(writer, &res)
	} else {
		res := []map[string]interface{}{}
		for i := 0; i < len(chathst.content); i++ {
			t1 := map[string]interface{}{"sender": chathst.sender[i], "receiver": chathst.receiver[i],
				"content": chathst.content[i], "date": chathst.date[i], "file": chathst.file[i], "filesrc": chathst.filesrc[i]}
			res = append(res, t1)
		}
		js, err := json.Marshal(res)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(js)
	}
}
