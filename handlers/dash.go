package handlers

// import (
// 	"encoding/json"
// 	"net/http"

// 	"go.mongodb.org/mongo-driver/bson"
// 	"juno.api/internal"
// )

// func (a *App) Dash(w http.ResponseWriter, r *http.Request) {
// 	claims, ok := Verify(w, r)
// 	if !ok {
// 		a.ClientError(w, http.StatusUnauthorized)
// 		return
// 	}
// 	userId := claims["user_id"]

// 	posts, err := internal.Get[internal.Product](r.Context(), &a.Database, "products", bson.M{"user_id": userId})
// 	if err != nil {
// 		a.ServerError(w, "Dash", err)
// 		return
// 	}
// 	if !ok {
// 		http.Error(w, "No Posts", http.StatusExpectationFailed)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(posts)

// }
