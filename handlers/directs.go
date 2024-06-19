package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"juno.api/internal"
)

func (a *App) Directs(w http.ResponseWriter, r *http.Request) {
	claims, ok := Verify(w, r)
	if !ok {
		return
	}

	userId := claims["user_id"]

	directs, err := internal.Get[internal.Direct](r.Context(), &a.Database, "directs", bson.M{
		"user_id": userId,
	})
	if err != nil {
		a.ServerError(w, "Directs", err)
		return
	}

	json.NewEncoder(w).Encode(directs)
}

func (a *App) Chats(w http.ResponseWriter, r *http.Request) {
	claims, ok := Verify(w, r)
	if !ok {
		return
	}

	userId := claims["user_id"].(string)

	directs, err := internal.Get[internal.Direct](r.Context(), &a.Database, "directs", bson.M{
		"$or": []bson.M{
			{"sender": userId},
			{"receiver": userId},
		},
	})

	if err != nil {
		a.ServerError(w, "Directs", err)
		return
	}

	users := map[string]internal.RenderedChat{}
	rendered := []internal.RenderedChat{}

	for _, direct := range directs {
		var data internal.RenderedDirect
		if direct.Sender == userId {
			data.Sent = true
		} else {
			data.Sent = false
		}

		// TODO : Add time to direct
		data.Content = direct.Content
		data.TimeSent = "4:45 PM"

		user := ""

		if data.Sent {
			user = direct.Receiver
		} else {
			user = direct.Sender
		}

		if _, ok := users[user]; !ok {
			var userData internal.User
			ok, err := a.Database.Get(r.Context(), "users", bson.M{"id": user}, &userData)
			if err != nil {
				a.ServerError(w, "Directs", err)
				return
			}
			if !ok {
				continue
			}

			users[user] = internal.RenderedChat{
				Name:     userData.Name,
				Username: userData.Username,
				Avatar:   userData.Avatar,
				Messages: []internal.RenderedDirect{},
			}

			v := users[user]
			v.Messages = append(v.Messages, data)
			users[user] = v
		} else {
			v := users[user]
			v.Messages = append(v.Messages, data)
			users[user] = v
		}

	}

	for key, data := range users {
		log.Println("key =", key)
		log.Println("final data =", data)
		rendered = append(rendered, data)
	}

	json.NewEncoder(w).Encode(rendered)
}
