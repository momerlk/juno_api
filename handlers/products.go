package handlers

import (
	"context"
	"fmt"
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
	
	filter := &FilterResponse{}
	for _ , brand := range data {
		label := CapitalizeWords(strings.ReplaceAll(brand.(string) , "_" , " "))
		filter.Brands = append(filter.Brands, FilterValue{Label : label , Value : brand.(string)})
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

	// Iterate over the cursor and decode each product
	if err = cursor.All(r.Context(), &products); err != nil {
		// Handle error
		log.Println("Error decoding products:", err)
		return
	}

	// TODO : Retrieve all the products data from their ids

	json.NewEncoder(w).Encode(products)
}

func (a *App) Recommend(n int) ([]internal.Product, error) {
	// get a cursor over the aggregation of products
	cur, err := a.Database.Collection(productsColl).Aggregate(
		context.TODO(),
		bson.A{bson.M{"$sample": bson.M{"size": n}}},
	)
	if err != nil {
		return nil, err
	}

	var results []internal.Product
	err = cur.All(context.TODO(), &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (a *App) Products(w http.ResponseWriter, r *http.Request) {
	_, ok := internal.Verify(w, r)
	if !ok {
		return
	}

	query_N := r.URL.Query().Get("n")
	n, err := strconv.Atoi(query_N)
	if err != nil {
		http.Error(w, "Query parameter n is not a valid integer", http.StatusBadRequest)
		return
	}

	// get a cursor over the aggregation of products
	cur, err := a.Database.Collection(productsColl).Aggregate(
		r.Context(),
		bson.A{bson.M{"$sample": bson.M{"size": n}}},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not get aggregation of products, err : %v", err.Error()), http.StatusInternalServerError)
		return
	}

	var results []internal.Product
	err = cur.All(r.Context(), &results)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not get aggregation of products, err : %v", err.Error()), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(results)
}

// TODO : Group items by brand
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

	productsByVendor := make(map[string][]internal.Product)

	productIds := []string{}	

	for _, action := range actions {
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

	json.NewEncoder(w).Encode(productsByVendor)
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

	limitStage := bson.D{{Key: "$limit", Value: 50}}

	// Perform the search
	collection := a.Database.Collection(productsColl)
	cursor, err := collection.Aggregate(r.Context(), mongo.Pipeline{query, limitStage})
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

	var body internal.ActionQuery
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Failed to decode body", http.StatusBadRequest)
		return
	}

	products, err := a.RecommendWithQuery(internal.Action{Query: body}, 50)
	if err != nil {
		http.Error(w, "Failed to get query", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(products)
}
