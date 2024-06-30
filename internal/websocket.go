package internal

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func ParseTokenString(tokenString string) (jwt.MapClaims , bool){
	var tokenClaims jwt.MapClaims

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		key := Getenv("JWT_KEY")
		return []byte(key) , nil
	})
	if err != nil {
		return nil , false
	}

	// Check if the token is valid and not expired
	if claims , ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		return nil , false
	} else {
		tokenClaims = claims
	}

	return tokenClaims , true;
}

func Verify(w http.ResponseWriter , r *http.Request) (jwt.MapClaims , bool) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Authorization token missing", http.StatusUnauthorized)
		return nil , false
	}

	var tokenClaims jwt.MapClaims

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		key := Getenv("JWT_KEY")
		return []byte(key) , nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return nil , false
	}

	// Check if the token is valid and not expired
	if claims , ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		http.Error(w, "Invalid token (expired)", http.StatusUnauthorized)
		return nil , false
	} else {
		tokenClaims = claims
	}


	return tokenClaims , true
}


// represents a websocket endpoint
type WebSocket struct {
	epoller 		*Epoll
}

func (socket *WebSocket)Init(mux *http.ServeMux, endpoint string, onRead func(ws *WebSocket, conn *WSConnection, data []byte) error) {
	// Start epoll
	var err error
	epoller, err := MkEpoll()
	if err != nil {
		panic(err)
	}

	socket.epoller = epoller;

	go socket.Start(onRead)

	mux.Handle(endpoint , socket)
}

func (socket *WebSocket) Send(userId string , data []byte) (bool , error) {
	conn , ok := socket.epoller.Writers[userId]
	if !ok {
		return false , nil
	}

	conn.Lock.Lock()
	defer conn.Lock.Unlock()

	err := wsutil.WriteServerText(conn.Conn , data)
	if err != nil {
		log.Printf("error on sent = %v\n" , err)
		return false , err
	}



	return true , nil
}

func (socket *WebSocket) Start(onRead func(websocket *WebSocket, conn *WSConnection , data []byte) error ){
	log.Println("started epolling")
	for {
		connections, err := socket.epoller.Wait()
		
		if err != nil {
			if strings.Contains(err.Error() , "interrupted system call"){
				continue
			}
			log.Printf("Failed to epoll wait %v", err)
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}
			if msg, _, err := wsutil.ReadClientData(conn.NetConn); err != nil {
				if err := socket.epoller.Remove(conn); err != nil {
					log.Printf("Failed to remove %v", err)
				}
				conn.NetConn.Close()
			} else {
				err = onRead(socket , conn , msg)
				if err != nil { 
					log.Printf("failed socket.onRead , connection user id = %v, err = %v\n" , conn.UserId , err)
					conn.NetConn.Close()
					continue
				}
			}
		}
	}
}

func (socket *WebSocket) ServeHTTP(w http.ResponseWriter , r *http.Request){
	// did not use internal.Verify because websockets don't support headers on clients such as react native
	token := r.URL.Query().Get("token");
	claims , ok := ParseTokenString(token)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	userId := claims["user_id"].(string)


	// Upgrade connection
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		log.Printf("failed to upgrade websocket , err = %v\n" , err)
		return
	}
	if err := socket.epoller.Add(&WSConnection{NetConn : conn, UserId: userId}); err != nil {
		log.Printf("Failed to add connection %v", err)
		conn.Close()
	}
}

type WSMessage struct {
	Message			string 			`json:"message" bson:"message"`
	Status 			int 			`json:"status" bson:"status"`
}

// TODO : Add error handling
func (socket *WebSocket) Message(conn *WSConnection,status int,message string){
	data , _ := json.Marshal(&WSMessage{
		Message : message,
		Status : status,
	})
	socket.Send(conn.UserId , data)
}