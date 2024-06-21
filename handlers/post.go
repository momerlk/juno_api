package handlers

// import (
// 	"context"
// 	"encoding/json"
// 	"log"

// 	jwt "github.com/golang-jwt/jwt/v5"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"juno.api/internal"

// 	"net/http"
// )

// const productsColl = "products"

// func (a *App) CreatePost(w http.ResponseWriter, r *http.Request) {

// 	if r.Method != http.MethodPost {
// 		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
// 		return
// 	}

// 	var claims jwt.MapClaims
// 	var ok bool
// 	if claims, ok = Verify(w, r); !ok {
// 		return
// 	}

// 	userId := claims["user_id"].(string)

// 	var data internal.Product

// 	var user internal.User
// 	ok, err := a.Database.Get(r.Context(), "users", bson.M{"id": userId}, &user)
// 	if !ok || err != nil {
// 		log.Printf("no user found, err = %v", err)
// 		return
// 	}

// 	err = json.NewDecoder(r.Body).Decode(&data)
// 	if err != nil {
// 		log.Println(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	data.Id = internal.GenerateId()
// 	data.UserId = user.Id
// 	data.Username = user.Username
// 	data.Avatar = user.Avatar

// 	err = a.Database.Store(context.TODO(), productsColl, data)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Write([]byte("Successfully posted data"))

// }
