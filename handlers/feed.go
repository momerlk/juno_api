package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"juno.api/internal"
)

const historiesColl = "histories"
const actionsColl = "actions"


// websocket handler for directs
func (a *App) WSFeed(ws *internal.WebSocket, conn *internal.WSConnection, data []byte) (err error) {

	var action internal.Action
	err = json.Unmarshal(data, &action)
	if err != nil {
		return err
	}

	log.Printf("feed served to user id = %v, action = %v\n" , conn.UserId , action)


	switch action.ActionType {
	case "open" :
		a.handleOpen(ws , conn, action);
	case "undo":
		a.handleUndo(ws , conn);
	default :
		a.handleSwipes(ws , conn, action);
	}

	return err
}

// handle undo action
func (a *App) handleUndo(
	ws *internal.WebSocket,
	conn *internal.WSConnection,
){
	// retrieve user's recommendation history
	var userHistory internal.UserHistory
	found , err := a.Database.Get(
		context.TODO() , 
		historiesColl , 
		bson.M{"user_id" : conn.UserId} , 
		&userHistory,
	)
	if err != nil{ // internal error in database
		ws.Message(
			conn ,
			http.StatusInternalServerError , 
			"Failed to retrieve user history",
		);
		return;
	}
	if !found { // no user history
		ws.Message(
			conn,
			http.StatusBadRequest,
			"First send 'open' message to websocket",
		)
		return;
	}

	userHistory.Index = userHistory.Index - 1;

	// products which need to be retrieved from database
	toFetch := userHistory.Products;
	var toFetchLen int
	if len(toFetch) <= 3 { toFetchLen = len(toFetch) } else { toFetchLen = 3 }


	// Update user history
	coll := a.Database.Collection(historiesColl)
	coll.FindOneAndReplace(
		context.TODO() , 
		bson.M{"user_id" : conn.UserId} , 
		userHistory,
	);

	// products to display to user
	var products []internal.Product
	for i := userHistory.Index;i < toFetchLen;i++ {
		
		var product internal.Product
		productId := toFetch[i];

		a.Database.Get(
			context.TODO(), 
			"products",
			bson.M{"product_id" : productId},
			&product,
		)

		products = append(products , product)
	}

	data , err := json.Marshal(products)
	if err != nil {
		ws.Message(
			conn,
			http.StatusInternalServerError,
			"Failed to encode products into json",
		)
		return;
	}


	ws.Send(conn.UserId , data);
}


// handles actions with action type "open"
func (a *App) handleOpen(
	ws *internal.WebSocket, 
	conn *internal.WSConnection,
	action internal.Action,
){

	// retrieve user's recommendation history
	var userHistory internal.UserHistory
	found , err := a.Database.Get(
		context.TODO() , 
		historiesColl , 
		bson.M{"user_id" : conn.UserId} , 
		&userHistory,
	)
	if err != nil { // internal error in database
		log.Println("WSFeed : failed to retrieve user history, userid =" , conn.UserId)
		ws.Message(
			conn ,
			http.StatusInternalServerError , 
			"Failed to retrieve user history",
		);
		return;
	}


	if !found { // IF FIRST TIME USER
		log.Println("WSFeed : first time user, userid =" , conn.UserId)
		recProducts , err := a.Recommend(3);
		if err != nil {
			ws.Message(
				conn,
				http.StatusInternalServerError,
				"Failed to recommend products",
			)
			return;
		}

		// store an array of the product ids of the recommended products
		var productIds []string 
		for i := 0;i < len(recProducts);i++ {
			productIds = append(productIds, recProducts[i].ProductID)
		}

		// user history with recommnded products
		newHistory := &internal.UserHistory{
			UserID : conn.UserId,
			Products : productIds,
			Index : 0,	
		}

		err = a.Database.Store(context.TODO() , historiesColl , newHistory)
		if err != nil {
			ws.Message(
				conn, 
				http.StatusInternalServerError,
				"Failed to store user history",
			)
			return;
		}

		// products to display to user
		var products []internal.Product
		for i := 0;i < len(newHistory.Products);i++ {
			
			var product internal.Product
			productId := newHistory.Products[i];

			a.Database.Get(
				context.TODO(), 
				"products",
				bson.M{"product_id" : productId},
				&product,
			)

			products = append(products , product)
		}


		// TODO : update send to do this automatically
		data , err := json.Marshal(products)
		if err != nil {
			ws.Message(
				conn,
				http.StatusInternalServerError,
				"Failed to encode products into json",
			)
			return;
		}
		sent , err := ws.Send(conn.UserId , data);
		if err != nil {
			log.Printf("Failed to write to websocket , err = %v\n",err)
			return;
		}
		if !sent {
			log.Printf("Connection closed failed to send")
		}


	} else { // NOT A FIRST TIME USER
		userHistory.Index += 1; // update the index 
		log.Println("WSFeed : not first time user, userid =" , conn.UserId)

		// if there are more than 3 items in user history after index return
		notSeen := len(userHistory.Products) - userHistory.Index
		if notSeen > 2 {
			// Update user history
			histColl := a.Database.Collection(historiesColl)
			histColl.FindOneAndReplace(
				context.TODO() , 
				bson.M{"user_id" : conn.UserId} , 
				userHistory,
			);
			// This gives error when the user opens the socket and wants to see new products

			// store an array of the product ids of the recommended products
			// products which need to be retrieved from database
			toFetch := userHistory.Products;

			// Update user history
			coll := a.Database.Collection(historiesColl)
			coll.FindOneAndReplace(
				context.TODO() , 
				bson.M{"user_id" : conn.UserId} , 
				userHistory,
			);

			// products to display to user
			var products []internal.Product
			for i := userHistory.Index;i < len(toFetch);i++ {
				
				var product internal.Product
				productId := toFetch[i];

				a.Database.Get(
					context.TODO(), 
					"products",
					bson.M{"product_id" : productId},
					&product,
				)

				products = append(products , product)
			}

			// shorten products to less or equal to three items
			if len(products) >= 3{
				products = products[0:3]
			}
			data , err := json.Marshal(products)
			if err != nil {
				ws.Message(
					conn,
					http.StatusInternalServerError,
					"Failed to encode products into json",
				)
				return;
			}


			ws.Send(conn.UserId , data);

			return;
		}

		// by using toRecommend reduces the number of operations
		recProducts , err := a.Recommend(2);
		if err != nil {
			ws.Message(
				conn,
				http.StatusInternalServerError,
				"Failed to recommend products",
			)
			return;
		}

		// store an array of the product ids of the recommended products
		productIds := userHistory.Products
		for i := 0;i < len(recProducts);i++ {
			productIds = append(productIds, recProducts[i].ProductID)
		}


		// products which need to be retrieved from database
		toFetch := userHistory.Products;

		userHistory.Products = productIds; // updated ids


		// Update user history
		coll := a.Database.Collection(historiesColl)
		coll.FindOneAndReplace(
			context.TODO() , 
			bson.M{"user_id" : conn.UserId} , 
			userHistory,
		);

		// products to display to user
		var products []internal.Product
		for i := userHistory.Index;i < len(toFetch);i++ {
			
			var product internal.Product
			productId := toFetch[i];

			a.Database.Get(
				context.TODO(), 
				"products",
				bson.M{"product_id" : productId},
				&product,
			)

			products = append(products , product)
		}

		products = append(products , recProducts...)
		// shorten products to less or equal to three items
		if len(products) >= 3{
			products = products[0:3]
		}
		data , err := json.Marshal(products)
		if err != nil {
			ws.Message(
				conn,
				http.StatusInternalServerError,
				"Failed to encode products into json",
			)
			return;
		}


		ws.Send(conn.UserId , data);
	}

	

}


// for handling swipes
func (a *App) handleSwipes(
	ws *internal.WebSocket, 
	conn *internal.WSConnection,
	action internal.Action,
){

	data := &internal.Action{
		UserID : conn.UserId,
		ProductID: action.ProductID,
		ActionType : action.ActionType,
		ActionID: uuid.NewString(),
		ActionTimestamp: time.Now().String(),
	}

	// retrieve user's recommendation history
	var userHistory internal.UserHistory
	found , err := a.Database.Get(
		context.TODO() , 
		historiesColl , 
		bson.M{"user_id" : conn.UserId} , 
		&userHistory,
	)
	if err != nil{ // internal error in database
		ws.Message(
			conn ,
			http.StatusInternalServerError , 
			"Failed to retrieve user history",
		);
		return;
	}
	if !found { // no user history
		ws.Message(
			conn,
			http.StatusBadRequest,
			"First send 'open' message to websocket",
		)
		return;
	}


	err = a.Database.Store(context.TODO() , actionsColl , data)
	if err != nil {
		ws.Message(
			conn,
			http.StatusInternalServerError,
			"Failed to save user action",
		)
		return;
	}

	userHistory.Index += 1; // update the index 


	// if there are more than 3 items in user history after index return
	notSeen := len(userHistory.Products) - userHistory.Index
	if notSeen > 2 {
		// Update user history
		coll := a.Database.Collection(historiesColl)
		coll.FindOneAndReplace(
			context.TODO() , 
			bson.M{"user_id" : conn.UserId} , 
			userHistory,
		);

		return;
	}

	// by using toRecommend reduces the number of operations
	recProducts , err := a.Recommend(2);
	if err != nil {
		ws.Message(
			conn,
			http.StatusInternalServerError,
			"Failed to recommend products",
		)
		return;
	}

	// store an array of the product ids of the recommended products
	productIds := userHistory.Products
	for i := 0;i < len(recProducts);i++ {
		productIds = append(productIds, recProducts[i].ProductID)
	}


	// products which need to be retrieved from database
	toFetch := userHistory.Products;

	userHistory.Products = productIds; // updated ids


	// Update user history
	coll := a.Database.Collection(historiesColl)
	coll.FindOneAndReplace(
		context.TODO() , 
		bson.M{"user_id" : conn.UserId} , 
		userHistory,
	);

	// products to display to user
	var products []internal.Product
	for i := userHistory.Index;i < len(toFetch);i++ {
		
		var product internal.Product
		productId := toFetch[i];

		a.Database.Get(
			context.TODO(), 
			"products",
			bson.M{"product_id" : productId},
			&product,
		)

		products = append(products , product)
	}

	products = append(products , recProducts...)
	// shorten products to less or equal to three items
	if len(products) >= 3{
		products = products[0:3]
	}
	toSend , err := json.Marshal(products)
	if err != nil {
		ws.Message(
			conn,
			http.StatusInternalServerError,
			"Failed to encode products into json",
		)
		return;
	}


	ws.Send(conn.UserId , toSend);
}

