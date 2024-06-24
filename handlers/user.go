package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"encoding/json"

	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"juno.api/internal"

	"net/http"
)

const usersColl = "users"

func GenerateToken(UserId string) (string, error){
	secret := internal.Getenv("JWT_KEY")
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":    UserId,
			"session_id": internal.GenerateId(),
			"exp":        time.Now().Add(4 * time.Hour).Unix(),
		})

	tokenString, err := token.SignedString([]byte(secret))
	return tokenString , err;
}

func (a *App) SignUp(w http.ResponseWriter, r *http.Request) {
	var body internal.User
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		a.ServerError(w, "Sign Up", err)
		return
	}

	log.Println("password =", body.Password)

	body.Id = uuid.NewString()

	hashed, err := internal.HashAndSalt([]byte(body.Password))
	if err != nil {
		a.ServerError(w, "Sign Up", err)
		return
	}
	body.Password = hashed

	a.Database.Store(r.Context(), usersColl, body)

	w.Write([]byte("successfully registered user"))
}

type SignInBody struct {
	UsernameEmail string `json:"username_email" bson:"username_email"` // username or email
	Password      string `json:"password" bson:"password"`
}
type TokenResp struct {
	Token string `json:"token" bson:"token"`
}

func (a *App) SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var body SignInBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		a.ServerError(w, "Sign In", err)
		return
	}

	var user internal.User
	ok, err := a.Database.Get(r.Context(), usersColl, bson.M{"username": body.UsernameEmail}, &user)
	if err != nil {
		a.ServerError(w, "Sign In a.Database.Get()", err)
		return
	}
	if !ok {
		ok, err := a.Database.Get(r.Context(), usersColl, bson.M{"email": body.UsernameEmail}, &user)
		if !ok {
			a.ClientError(w, http.StatusUnauthorized)
			return
		}
		if err != nil {
			a.ServerError(w, "Sign In", err)
			return
		}
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)) == nil {
		log.Println("user is authenticated")

		tokenString, err := GenerateToken(user.Id);
		if err != nil {
			a.ServerError(w, "Sign In", err)
			return
		}

		user.Password = ""
		user.Id = ""

		err = json.NewEncoder(w).Encode(TokenResp{Token: tokenString})
		if err != nil {
			a.ServerError(w, "Sign In", err)
			return
		}

	} else {
		a.ClientError(w, http.StatusUnauthorized)
		return
	}

}

func Verify(w http.ResponseWriter, r *http.Request) (jwt.MapClaims, bool) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Authorization token missing", http.StatusUnauthorized)
		return nil, false
	}

	var tokenClaims jwt.MapClaims

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		key := internal.Getenv("JWT_KEY")
		return []byte(key), nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return nil, false
	}

	// Check if the token is valid and not expired
	if claims, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		http.Error(w, "Invalid token (expired)", http.StatusUnauthorized)
		return nil, false
	} else {
		tokenClaims = claims
	}

	return tokenClaims, true
}

func (a *App) Refresh(w http.ResponseWriter , r *http.Request){
	if r.Method != http.MethodGet {
		a.ClientError(w , http.StatusMethodNotAllowed);
		return
	}

	claims , ok := internal.Verify(w ,r);
	if !ok {
		return;
	}

	userId := claims["user_id"];

	token, err := GenerateToken(userId.(string))
	if err != nil {
		http.Error(w , "Failed to generate authentication token" , http.StatusInternalServerError);
		return;
	}

	err = json.NewEncoder(w).Encode(TokenResp{Token: token})
	if err != nil {
		a.ServerError(w, "Sign In", err)
		return
	}
}

// GET : Retrieve user details
func (a *App) Details(w http.ResponseWriter, r *http.Request) {
	claims, ok := internal.Verify(w, r)
	if !ok {
		return
	}

	userId := claims["user_id"]

	var user internal.User
	a.Database.Get(r.Context(), usersColl, bson.M{"id": userId}, &user)

	user.Password = ""
	user.Id = ""

	json.NewEncoder(w).Encode(user)
}

func (a *App) Liked(w http.ResponseWriter , r *http.Request){
	claims , ok := internal.Verify(w , r);
	if !ok {
		return
	}

	userId := claims["user_id"]

	var product []interface{}
	a.Database.Get(r.Context(), "products" , bson.M{"user_id" : userId} , &product)

	// TODO : Retrieve all the products data from their ids

	json.NewEncoder(w).Encode(product)
}

func (a *App) Recommend(n int) ([]internal.Product, error) {
	// get a cursor over the aggregation of products
	cur , err := a.Database.Collection("products").Aggregate(
		context.TODO(),
		bson.A{bson.M{"$sample": bson.M{"size": n}}},
	)
	if err != nil {
		return nil,  err
	}

	var results []internal.Product
	err = cur.All(context.TODO() , &results)
	if err != nil {
		return nil , err
	}

	return results , nil
}

func (a *App) Products(w http.ResponseWriter , r *http.Request){
	_ , ok := internal.Verify(w , r);
	if !ok {
		return
	}

	query_N := r.URL.Query().Get("n")
	n , err := strconv.Atoi(query_N)
	if err != nil {
		http.Error(w , "Query parameter n is not a valid integer" , http.StatusBadRequest);
		return
	}

	// get a cursor over the aggregation of products
	cur , err := a.Database.Collection("products").Aggregate(
		r.Context() , 
		bson.A{bson.M{"$sample": bson.M{"size": n}}},
	)
	if err != nil {
		http.Error(w , fmt.Sprintf("Could not get aggregation of products, err : %v" , err.Error()) , http.StatusInternalServerError)
		return
	}

	var results []internal.Product
	err = cur.All(r.Context() , &results)
	if err != nil {
		http.Error(w , fmt.Sprintf("Could not get aggregation of products, err : %v" , err.Error()) , http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(results)
}

// TODO : Group items by brand
func (a *App) Cart(w http.ResponseWriter , r *http.Request){}