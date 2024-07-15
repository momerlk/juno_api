package handlers

import (
	"log"
	"strconv"
	"strings"
	"unicode"

	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"juno.api/internal"

	"net/http"
)

const productsColl = "products"

type FilterValue struct {
	Image 				string 				`json:"image" bson:"image"`
	Label 				string 				`json:"label" bson:"label"`
	Value 				string 				`json:"value" bson:"value"`
}
type FilterResponse struct {
	Brands 				[]FilterValue 		`json:"brands" bson:"brands"`
}

func CapitalizeWords(s string) string {
	words := strings.Fields(s) // Split the string into words
	for i, word := range words {
		runes := []rune(word)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ") // Join the words back into a single string
}

// GET /filter
func (a *App) Filter(w http.ResponseWriter, r *http.Request) {
	// no need for verification in this field
	if r.Method != http.MethodGet {
		a.ClientError(w , http.StatusMethodNotAllowed);
		return;
	}

	// getting all the unique brand values in the database
	data , err := a.Database.Collection(productsColl).Distinct(r.Context() , "vendor" , bson.D{})
	if err != nil {
		http.Error(w , "Failed to get distinct brand values" , http.StatusInternalServerError);
		return
	}

	var brandData []internal.Brand
	cur , err := a.Database.Collection(brandsColl).Find(r.Context() , bson.M{})
	if err != nil {
		a.ServerError(w , "/filter" , err)
		return
	}
	err = cur.All(r.Context() , &brandData)
	if err != nil {
		a.ServerError(w , "/filter" , err)
		return
	}

	var images map[string]string = map[string]string{}
	for _ , brand := range brandData {
		images[brand.Name] = brand.Logo;
	}

	
	filter := &FilterResponse{}
	for _ , brand := range data {
		label := CapitalizeWords(strings.ReplaceAll(brand.(string) , "_" , " "))

		filter.Brands = append(filter.Brands, FilterValue{Image : images[brand.(string)], Label : label , Value : brand.(string)})
	}

	json.NewEncoder(w).Encode(filter);
}


// GET /liked
func (a *App) Liked(w http.ResponseWriter, r *http.Request) {
	claims, ok := internal.Verify(w, r)
	if !ok {
		return
	}

	userId := claims["user_id"]

	var actions []internal.Action
	cursor, err := a.Database.Collection(actionsColl).Find(
		r.Context(),
		bson.M{"user_id": userId, "action_type": internal.LikeAction},
	)
	if err != nil {
		log.Println("GET /liked error =", err)
		http.Error(w, "Failed to retrieve user actions", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(r.Context())

	err = cursor.All(r.Context(), &actions)
	if err != nil {
		log.Println("GET /liked error =", err)
		http.Error(w, "Failed to retrieve user actions", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(r.Context())

	var products []internal.Product
	var productIDs []string

	// Collect all product IDs from the actions
	for _, action := range actions {
		productIDs = append(productIDs, action.ProductID)
	}

	// Fetch all products at once using the $in operator
	filter := bson.M{"product_id": bson.M{"$in": productIDs}}
	cursor, err = a.Database.Collection(productsColl).Find(r.Context(), filter)
	if err != nil {
		// Handle error
		log.Println("Error fetching products:", err)
		return
	}
	defer cursor.Close(r.Context())

	// Iterate over the cursor and decode each product
	if err = cursor.All(r.Context(), &products); err != nil {
		// Handle error
		log.Println("Error decoding products:", err)
		return
	}

	// TODO : Retrieve all the products data from their ids

	json.NewEncoder(w).Encode(products)
}


func (a *App) Products(w http.ResponseWriter, r *http.Request) {
	claims, ok := internal.Verify(w, r)
	if !ok {
		return
	}
	userId := claims["user_id"]

	query_N := r.URL.Query().Get("n")
	n, err := strconv.Atoi(query_N)
	if err != nil {
		http.Error(w, "Query parameter n is not a valid integer", http.StatusBadRequest)
		return
	}

	results , err := a.Recommend(userId.(string) , n)
	if err != nil {
		log.Println("recommendations system error =" , err)
		http.Error(w , "Failed to get recommendations internally" , http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(results)
}

type CartItem struct {
	Vendor 					string 					`json:"vendor" bson:"vendor"`
	Items 					[]internal.Product		`json:"items" bson:"items"`
}
func (a *App) Cart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.ClientError(w, http.StatusMethodNotAllowed)
		return
	}

	claims, ok := internal.Verify(w, r)
	if !ok {
		return
	}
	userId := claims["user_id"]

	
	actions, err := internal.Get[internal.Action](
		r.Context(), &a.Database, actionsColl,
		bson.M{"user_id": userId, "action_type": internal.AddToCartAction},
	)
	if err != nil {
		a.ServerError(w, "CART", err) // TODO : add error strings to server error
		return
	}

	deletedActions, err := internal.Get[internal.Action](
		r.Context(), &a.Database, actionsColl,
		bson.M{"user_id": userId, "action_type": internal.DeletedFromCartAction},
	)
	if err != nil {
		a.ServerError(w, "CART", err) // TODO : add error strings to server error
		return
	}

	productsByVendor := make(map[string][]internal.Product)

	productIds := []string{}	

	main : 
		for _, action := range actions {
			for _ , deleted := range deletedActions{
				if deleted.ProductID == action.ProductID{
					continue main
				}
			}	
			productIds = append(productIds, action.ProductID)
		}

	cursor , err := a.Database.Collection(productsColl).Find(
		r.Context() , 
		bson.M{"product_id" : bson.M{"$in" : productIds}},
	)
	if err != nil {
		http.Error(w , "/cart failed to get products" , http.StatusInternalServerError)
		return
	}
	defer cursor.Close(r.Context())

	var products []internal.Product
	err = cursor.All(r.Context() , &products);
	if err != nil {
		http.Error(w , "/cart failed to decode products" , http.StatusInternalServerError)
		return
	}

	for _ , product := range products {
		arr := productsByVendor[product.Vendor]
		arr = append(arr,product)
		productsByVendor[product.Vendor]  = arr
	}

	items := []CartItem{}
	for vendor , products := range productsByVendor {
		items = append(items , CartItem{
			Vendor: vendor,
			Items: products,
		})
	}

	json.NewEncoder(w).Encode(items)
}

func (a *App) SearchProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.ClientError(w, http.StatusMethodNotAllowed)
		return
	}

	queryString := r.URL.Query().Get("q")
	if queryString == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	randomString := r.URL.Query().Get("random")
	randomize := false
	if randomString == "yes" {
		randomize = true
	}
	

	// Construct the query with fuzzy parameters
	query := bson.D{
		{Key: "$search", Value: bson.D{
			{Key: "index", Value: "aisearch"}, // Ensure this matches the index name
			{Key: "text", Value: bson.D{
				{Key: "query", Value: queryString},
				{Key: "path", Value: bson.D{
					{Key: "wildcard", Value: "*"},
				}},
			}},
		}},
	}

	n := 70
	query_N := r.URL.Query().Get("n")
	if query_N != "" {
		result, _ := strconv.Atoi(query_N)
		if result > 0 {
			n = result
		}
	}

	// TODO : change limit
	var pipeline []bson.D

	limitStage  := bson.D{{Key: "$limit", Value: n}}
	

	if randomize {
		limitStage  = bson.D{{Key: "$limit", Value: 1000}}
		sampleStage := bson.D{
			bson.E{
				Key: "$sample", Value: bson.D{
					{Key: "size", Value: n},
				},
			},
		}
		pipeline = mongo.Pipeline{query, limitStage, sampleStage}
	} else {
		pipeline = mongo.Pipeline{query, limitStage}
	}

	// Perform the search
	collection := a.Database.Collection(productsColl)
	cursor, err := collection.Aggregate(r.Context(), pipeline)
	if err != nil {
		log.Println("Failed to perform search, err =", err)
		http.Error(w, "Failed to perform search", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(r.Context())

	var products []internal.Product
	if err = cursor.All(r.Context(), &products); err != nil {
		http.Error(w, "Failed to parse search results", http.StatusInternalServerError)
		return
	}

	// Encode the result as JSON and write to response
	json.NewEncoder(w).Encode(products)
}

// query the products
func (a *App) QueryProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.ClientError(w, http.StatusMethodNotAllowed)
		return
	}

	claims , ok := internal.Verify(w,r)
	if !ok {
		return
	}
	userId := claims["user_id"]

	var body internal.ActionQuery
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Failed to decode body", http.StatusBadRequest)
		return
	}

	products, err := a.RecommendWithQuery(internal.Action{
		UserID: userId.(string),
		Query: body,
	}, 50)
	if err != nil {
		http.Error(w, "Failed to get query", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(products)
}