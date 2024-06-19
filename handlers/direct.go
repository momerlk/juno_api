package handlers

import (
	"context"
	"encoding/json"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"pakmaweshi.api/internal"
)

const directsColl =  "directs"

type WSDirectMeta struct {
	Received 			bool 			`json:"received" bson:"received"`
}

// websocket handler for directs
func (a *App) WSDirect(ws *internal.WebSocket , conn *internal.WSConnection , data []byte) (err error) {
	var direct internal.Direct

	err = json.Unmarshal(data , &direct)
	if err != nil {
		return err
	}

	direct.Id = internal.GenerateId()
	direct.Received = false


	var receiver internal.User
	ok , err := a.Database.Get(context.TODO() , "users" , bson.M{"username" : direct.Receiver} , &receiver)
	if err != nil {
		return err
	}
	if !ok {
		return err
	}

	direct.Receiver = receiver.Id
	direct.Sender = conn.UserId

	log.Printf("WSDirect : direct = %v\n" , direct)

	newData , err := json.Marshal(direct)
	if err != nil {
		return err
	}

	ok , err = ws.Send(receiver.Id , newData)
	if err != nil {
		return err
	}

	if ok {
		direct.Received = true
	} else {
		direct.Received = false
	}

	err = a.Database.Store(context.TODO() , directsColl , direct)
	if err != nil {
		return err
	}

	// meta , err := json.Marshal(WSDirectMeta{
	// 	Received: direct.Received,
	// })
	// if err != nil {
	// 	return err
	// }

	// ws.Send(direct.Sender , meta)

	return err
}


