package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"juno.api/internal"
)

func (a *App) PostAction(w http.ResponseWriter , r *http.Request){
	claims , ok := internal.Verify(w , r);
	if !ok {
		return
	}

	userId := claims["user_id"].(string)

	var action internal.Action
	err := json.NewDecoder(r.Body).Decode(&action)
	if err != nil {
		a.ServerError(w, "Post Action", err)
		return
	}

	// TODO : If action on product id already exists update it

	actionData := &internal.Action{
		UserID:          userId,
		ProductID:       action.ProductID,
		ActionType:      action.ActionType,
		ActionID:        uuid.NewString(),
		ActionTimestamp: time.Now().String(),
	}

	err = a.Database.Store(context.TODO(), actionsColl, actionData)
	if err != nil {
		a.ServerError(
			w,
			"POST Action (Failed to save user action)",
			err,
		)
		return
	} else {
		w.Write([]byte("successfully added action to database"))
	}

}


func (a *App) Brands(w http.ResponseWriter , r *http.Request){
	if r.Method != http.MethodGet {
		a.ClientError(w , http.StatusMethodNotAllowed)
		return
	}

	cursor , err := a.Database.Collection(brandsColl).Find(r.Context() , bson.M{})
	if err != nil {
		a.ServerError(w , "/brands" , err)
		return
	}
	defer cursor.Close(r.Context())

	var brands []internal.Brand
	err = cursor.All(r.Context() , &brands);
	if err != nil {
		a.ServerError(w , "/brands" , err)
		return
	}

	json.NewEncoder(w).Encode(brands)
}