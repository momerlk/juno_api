package handlers

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"juno.api/internal"
)

const actionsColl = "actions"

// FEED Options : query, filter, see product

// websocket handler for directs


func (a *App) RecommendWithQuery(action internal.Action, n int) ([]internal.Product, error) {
	//log.Println("filter =" , action.Query.Filter)
	//log.Println("text =" , action.Query.Text)

	// filter based query only
	if action.Query.Filter != nil && action.Query.Text == "" {

		// Construct the aggregation pipeline
		pipeline := bson.A{
			bson.M{"$match": action.Query.Filter}, // Add $match stage to filter by category
			bson.M{"$limit" : n}, // limit to n items
		}

		// Perform aggregation
		cur, err := a.Database.Collection(productsColl).Aggregate(
			context.TODO(),
			pipeline,
		)
		if err != nil {
			return nil, err
		}
		defer cur.Close(context.TODO())

		// Decode results into a slice of internal.Product
		var results []internal.Product
		err = cur.All(context.TODO(), &results)
		if err != nil {
			return nil, err
		}
		defer cur.Close(context.TODO())

		remainingProducts := n - len(results)
		if remainingProducts > 2 {
			recs, err := a.Recommend(action.UserID , remainingProducts)
			if err != nil {
				return nil, err
			}
			results = append(results, recs...)
		}

		return results, nil
	}

	if action.Query.Text != "" {
		// Construct the query with fuzzy parameters
		query := bson.D{
			{Key: "$search", Value: bson.D{
				{Key: "index", Value: "aisearch"}, // Ensure this matches your index name
				{Key: "text", Value: bson.D{
					{Key: "query", Value: action.Query.Text},
					{Key: "path", Value: bson.D{
						{Key: "wildcard", Value: "*"},
					}},
				}},
			}},
		}

		limitStage := bson.D{{Key: "$limit", Value: n}}

		// Perform the search
		collection := a.Database.Collection(productsColl)

		var cursor *mongo.Cursor
		var err error
		if action.Query.Filter == nil {
			cursor, err = collection.Aggregate(context.TODO(), mongo.Pipeline{
				query, limitStage,
			})
			if err != nil {
				return nil, err
			}

		} else {
			cursor, err = collection.Aggregate(
				context.TODO(),
				bson.A{query, bson.M{"$match": action.Query.Filter}, limitStage},
			)
			if err != nil {
				log.Println("cursor with filter error =", err)
				return nil, err
			}

		}
		defer cursor.Close(context.TODO())

		var products []internal.Product
		if err = cursor.All(context.TODO(), &products); err != nil {
			return nil, err
		}

		remainingProducts := n - len(products)
		if remainingProducts > 2 {
			recs, err := a.Recommend(action.UserID , remainingProducts)
			if err != nil {
				return nil, err
			}
			products = append(products, recs...)
		}

		return products, nil
	}

	// standard feed
	return a.Recommend(action.UserID , n)
}

