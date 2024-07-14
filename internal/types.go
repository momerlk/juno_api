package internal


type Brand struct {
	BrandID 				string 					`json:"brand_id" bson:"brand_id"`
	Name					string					`json:"name" bson:"name"`
	Logo					string 					`json:"logo" bson:"logo"`
	BaseURL					string					`json:"base_url" bson:"base_url"`
	Description 			string 					`json:"description" bson:"description"`	
}


const LikeAction = "like"
const DislikeAction = "dislike"
const AddToCartAction = "added_to_cart"
const DeletedFromCartAction = "deleted_from_cart"
const PurchaseAction = "purchase"

// Action represents an action performed by a user
type Action struct {
	UserID          	string      		`json:"user_id" bson:"user_id"`
	ActionID        	string      		`json:"action_id" bson:"action_id"`
	ActionType      	string      		`json:"action_type" bson:"action_type"`
	ActionTimestamp 	string      		`json:"action_timestamp" bson:"action_timestamp"`
	ProductID       	string      		`json:"product_id" bson:"product_id"`
	Query           	ActionQuery 		`json:"query" bson:"query"`
}

// for product just add "product_id" to filter
type ActionQuery struct {
	Text   				string      		`json:"text"`
	Filter 				interface{} 		`json:"filter" bson:"filter"`
}

// Action represents an action performed by a user
type UserHistory struct {
	UserID   			string   			`json:"user_id" bson:"user_id"`
	Products 			[]string 			`json:"products" bson:"products"`
	Index    			int      			`json:"index" bson:"index"`
}

// Product represents a product in the store
type Product struct {
    ProductID    string    `json:"product_id" bson:"product_id"`
    ProductURL   string    `json:"product_url" bson:"product_url"`
    ShopifyID    string    `json:"shopify_id" bson:"shopify_id"`
    Handle       string    `json:"handle" bson:"handle"`
    Title        string    `json:"title" bson:"title"`
    Vendor       string    `json:"vendor" bson:"vendor"`
    VendorTitle  string    `json:"vendor_title" bson:"vendor_title"`
    Category     string    `json:"category" bson:"category"`
    ProductType  string    `json:"product_type" bson:"product_type"`
    ImageURL     string    `json:"image_url" bson:"image_url"`
    Images       []string  `json:"images" bson:"images"`
    Description  string    `json:"description" bson:"description"`
    Price        int       `json:"price" bson:"price"`
    ComparePrice int       `json:"compare_price" bson:"compare_price"`
    Discount     int       `json:"discount" bson:"discount"`
    Currency     string    `json:"currency" bson:"currency"`
    Variants     []Variant `json:"variants" bson:"variants"`
    Options      []Option  `json:"options" bson:"options"`
    Tags         []string  `json:"tags" bson:"tags"`
    Available    bool      `json:"available" bson:"available"`
}

// Variant represents a variant of the product
type Variant struct {
    ID           string `json:"id" bson:"id"`
    Price        int    `json:"price" bson:"price"`
    Title        string `json:"title" bson:"title"`
    ComparePrice int    `json:"compare_price" bson:"compare_price"`
    Option1      string `json:"option1" bson:"option1"`
    Option2      string `json:"option2" bson:"option2"`
    Option3      string `json:"option3" bson:"option3"`
}

// Option represents an option for the product
type Option struct {
    Name     string   `json:"name" bson:"name"`
    Position int      `json:"position" bson:"position"`
    Values   []string `json:"values" bson:"values"`
}

type User struct {
	Id 				string 	`json:"id" bson:"id"` // user id

	Avatar 			string 	`json:"avatar" bson:"avatar"` // url of the avatar image file
	Age 			int		`json:"age" bson:"age"`
	Gender 			string 	`json:"gender" bson:"gender"`

	Name     		string 	`json:"name" bson:"name"`         // full name
	PhoneNumber   	string 	`json:"phone_number" bson:"phone_number"`     // phone number only +92
	Username 		string 	`json:"username" bson:"username"` // username

	Email    		string 	`json:"email" bson:"email"`       // email
	Password 		string 	`json:"password" bson:"password"` // password
}

type Direct struct {
	Id string `json:"id" bson:"id"` // message id

	Sender   string `json:"sender" bson:"sender"`     // sender's user id
	Receiver string `json:"receiver" bson:"receiver"` // receiver's user id
	Received bool   `json:"received" bson:"received"` // whether the message has been received or not

	Content    string `json:"content" bson:"content"`       // text content of the message
	Attachment string `json:"attachment" bson:"attachment"` // file id of the attachment
}

type RenderedDirect struct {
	Content  string `json:"content" bson:"content"`
	TimeSent string `json:"time_sent" bson:"time_sent"`
	Sent     bool   `json:"sent"`
}

type RenderedChat struct {
	Name     string           `json:"name"`
	Username string           `json:"username"`
	Avatar   string           `json:"avatar"`
	Messages []RenderedDirect `json:"messages"`
}
