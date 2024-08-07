package handlers

import (
	"log"
	"strings"

	"encoding/json"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	"juno.api/internal"

	"net/http"
)

const usersColl = "users"

func FmtPhoneNumber (param string) string {
	PhoneNumber := param
	PhoneNumber = strings.ReplaceAll(PhoneNumber , " " , "")

	after , _ := strings.CutPrefix(PhoneNumber , "+92")
	after, found := strings.CutPrefix(after , "0")
	if found {
		PhoneNumber = "+92" + after
	}

	return PhoneNumber
}

func (a *App) VerifyToken(w http.ResponseWriter , r *http.Request){
	_ , ok := internal.Verify(w,r)
	if ok {
		w.WriteHeader(http.StatusOK)
		return
	}
}

func (a *App) SignUp(w http.ResponseWriter, r *http.Request) {
	var body internal.User
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		a.ServerError(w, "Sign Up", err)
		return
	}


	body.Id = uuid.NewString()

	hashed, err := internal.HashAndSalt([]byte(body.Password))
	if err != nil {
		a.ServerError(w, "Sign Up", err)
		return
	}
	body.Password = hashed

	// remove all whitespace
	body.PhoneNumber = FmtPhoneNumber(body.PhoneNumber)

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

	// TODO : not more than 5 devices on one account

	var body SignInBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		a.ServerError(w, "Sign In", err)
		return
	}

	var user internal.User
	ok, err := a.Database.Get(r.Context(), usersColl, bson.M{"phone_number": body.UsernameEmail}, &user)
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

		tokenString, err := internal.GenerateToken(user.Id);
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

	token, err := internal.GenerateToken(userId.(string))
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
