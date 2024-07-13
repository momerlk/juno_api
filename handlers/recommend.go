package handlers

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"juno.api/internal"
)


func (a *App) Recommend(userId string, n int) ([]internal.Product , error){

	var actions []internal.Action
	cursor , err := a.Database.Collection(actionsColl).Find(context.TODO(), bson.M{"user_id" : userId})
	if err != nil {
		return nil , err;
	}
	defer cursor.Close(context.TODO())

	err = cursor.All(context.TODO() , &actions)
	if err != nil {
		return nil , err
	}

	productIds := []string{}
	actionsMap := map[string]string{}
	for _ , action := range actions {
		productIds = append(productIds, action.ProductID)
		actionsMap[action.ProductID] = action.ActionType
	}

	var products []internal.Product
	cursor , err = a.Database.Collection(productsColl).Find(context.TODO() , bson.M{"product_id" : bson.M{"$in" : productIds}})
	if err != nil {
		return nil , err
	}
	defer cursor.Close(context.TODO())

	err = cursor.All(context.TODO() , &products);
	if err != nil {
		return nil , err;
	}
	

	// Sample user interaction arrays (replace with actual data)
    liked := []internal.Product{}
    // addedToCart := []internal.Product{}
    // purchased := []internal.Product{}
    disliked := []internal.Product{}
	for _ , product := range products {
		switch actionsMap[product.ProductID] {
		case internal.LikeAction:
			liked = append(liked , product)
		case internal.DislikeAction : 
			disliked = append(disliked , product)
		// case internal.AddToCartAction : 
		// 	addedToCart = append(addedToCart, product)
		// case internal.PurchaseAction : 
		// 	purchased = append(purchased, product)
		default : 
			continue	
		}
	}

	dislikedQuery :=  bson.D{
        {Key: "$search", Value: bson.D{
            {Key: "index", Value: "aisearch"}, 
            {Key: "moreLikeThis", Value: bson.D{
                {Key: "like", Value: disliked},
            }},
        }},
    }
	
	coll := a.Database.Collection(productsColl)
	// Execute the query
    cursor, err = coll.Aggregate(context.TODO(), 
		mongo.Pipeline{
			dislikedQuery,
			bson.D{{Key: "$limit", Value: 100}}, // Limit the results to 100
    	},
	)
    if err != nil {
        return nil , err
    }
    defer cursor.Close(context.TODO())
	// Process the results
    var topDisliked []internal.Product // top 100 disliked products
    if err = cursor.All(context.TODO(), &topDisliked); err != nil {
        return nil , err
    }
	dislikedIds := []string{}
	for _ , product := range topDisliked {
		dislikedIds = append(dislikedIds, product.ProductID)
	}


	// Define the query
    query := bson.D{
        {Key: "$search", Value: bson.D{
            {Key: "index", Value: "aisearch"}, 
            {Key: "moreLikeThis", Value: bson.D{
                {Key: "like", Value: liked},
            }},
        }},
    }

    // Execute the query
    cursor, err = coll.Aggregate(context.TODO(), 
		mongo.Pipeline{
			query,
			bson.D{{Key: "$match", Value: bson.D{
				{Key: "product_id", Value: bson.D{
					{Key: "$nin", Value: dislikedIds},
				}},
			}}}, // remove disliked products
			bson.D{{Key: "$match", Value: bson.D{
				{Key: "product_id", Value: bson.D{
					{Key: "$nin", Value: productIds},
				}},
			}}}, // remove seen products
			bson.D{{Key: "$limit", Value: n}}, // Limit the results to n
    	},
	)
    if err != nil {
        return nil , err
    }
    defer cursor.Close(context.TODO())


    // Process the results
    var results []internal.Product
    if err = cursor.All(context.TODO(), &results); err != nil {
        return nil , err
    }

	return results , nil;
}