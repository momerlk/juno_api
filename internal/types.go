package internal

import (
)

// Action represents an action performed by a user
type Action struct {
    UserID         	    string          `json:"user_id" bson:"user_id"`
    ActionID       	    string          `json:"action_id" bson:"action_id"`
    ActionType     	    string          `json:"action_type" bson:"action_type"`
    ActionTimestamp     string          `json:"action_timestamp" bson:"action_timestamp"`
    ProductID      	    string          `json:"product_id" bson:"product_id"`
    DataType            string          `json:"data_type" bson:"data_type"` 
    Query               ActionQuery     `json:"query" bson:"query"` 
}

// for product just add "product_id" to filter
type ActionQuery struct {
    Text                string          `json:"text"`
    Filter              interface{}     `json:"filter" bson:"filter"`
}

// Action represents an action performed by a user
type UserHistory struct {
    UserID   string   `json:"user_id" bson:"user_id"`
    Products []string `json:"products" bson:"products"`
    Index    int      `json:"index" bson:"index"`
}


// Product represents the structure of the JSON document
type Product struct {
    ProductID    string              `json:"product_id" bson:"product_id"`
    ProductURL   string              `json:"product_url" bson:"product_url"`
    ShopifyID    string           	 `json:"shopify_id" bson:"shopify_id"`
    Handle       string              `json:"handle" bson:"handle"`
    Title        string              `json:"title" bson:"title"`
    Vendor       string              `json:"vendor" bson:"vendor"`
    Category     string              `json:"category" bson:"category"`
    ImageURL     string              `json:"image_url" bson:"image_url"`
    Description  string              `json:"description" bson:"description"`
    BodyHTML     string              `json:"body_html" bson:"body_html"`
    Price        uint64              `json:"price" bson:"price"`
    Currency     string              `json:"currency" bson:"currency"`
    Options      []ProductOption     `json:"options" bson:"options"`
    Tags         []string            `json:"tags" bson:"tags"`
    Available    bool                `json:"available" bson:"available"`
}

// ProductOption represents a product option
type ProductOption struct {
    Name     string   `json:"name" bson:"name"`
    Position int      `json:"position" bson:"position"`
    Values   []string `json:"values" bson:"values"`
}

type User struct {
	Id 				string 					`json:"id" bson:"id"` // user id 

	Avatar			string 					`json:"avatar" bson:"avatar"` // url of the avatar image file

	Name 			string					`json:"name" bson:"name"` // full name
	Number 			string 					`json:"number" bson:"number"` // phone number only +92
	Username		string 					`json:"username" bson:"username"` // username

	Email			string 					`json:"email" bson:"email"` // email
	Password		string 					`json:"password" bson:"password"`	 // password
}

type Direct struct {
	Id 				string 					`json:"id" bson:"id"` 					// message id

	Sender 			string 					`json:"sender" bson:"sender"` 			// sender's user id
	Receiver		string 					`json:"receiver" bson:"receiver"` 		// receiver's user id
	Received 		bool					`json:"received" bson:"received"` 		// whether the message has been received or not

	Content 		string 					`json:"content" bson:"content"` 	   // text content of the message
	Attachment 		string 					`json:"attachment" bson:"attachment"` // file id of the attachment
}

type RenderedDirect struct {
	Content 		string 					`json:"content" bson:"content"`
	TimeSent 		string 					`json:"time_sent" bson:"time_sent"`
	Sent 			bool					`json:"sent"`
}

type RenderedChat struct {
	Name 			string					`json:"name"`
	Username 		string					`json:"username"`
	Avatar 			string					`json:"avatar"`
	Messages 		[]RenderedDirect 		`json:"messages"`
}