package internal

import (
	"context"
	"bytes"
	"log"
	"fmt"
	"io"
	"os"


	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	mongoDB 				*mongo.Database
}


func (d *Database) Init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}
	dbName := os.Getenv("MONGODB_DBNAME")



	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	
	
	d.mongoDB = client.Database(dbName)
}

func (d *Database) Collection(name string , opts ...*options.CollectionOptions) *mongo.Collection {
	return d.mongoDB.Collection(name , opts...);
}

func (d *Database) Store(ctx context.Context , collName string , data interface{}) (error){
	coll := d.mongoDB.Collection(collName)

	_ , err := coll.InsertOne(ctx , data)

	return err 
}

func (d *Database) Get(ctx context.Context , collName string , filter interface{} , data interface{}) (bool , error){
	coll := d.mongoDB.Collection(collName)

	res := coll.FindOne(ctx , filter)
	if res.Err() == mongo.ErrNoDocuments {
		return false , nil
	}
	err := res.Decode(data)
	if err != nil {
		return false , err
	}

	return true , nil
}

func Get[T any](ctx context.Context, d *Database  , collName string , filter interface{}) ([]T , error){
	var data []T
	coll := d.mongoDB.Collection(collName)

	cur , err := coll.Find(ctx , filter)
	if cur.Err() == mongo.ErrNoDocuments {
		return nil , nil
	} 
	if err != nil {
		return data , err
	}


	for cur.Next(ctx) {
		var item T
		err = cur.Decode(&item)
		if err != nil {
			return data , err
		}
		data = append(data , item)
	}

	return data , nil
}


func (d *Database) StoreJPG(id string , data []byte) error {
	opts := options.GridFSBucket().SetName("images")
	bucket , err := gridfs.NewBucket(d.mongoDB ,opts)
	if err != nil {
		return err
	}

	_ , err = bucket.UploadFromStream(fmt.Sprintf("%v.jpg" , id) , bytes.NewReader(data))
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) GetJPG(id string , w io.Writer) error {
	opts := options.GridFSBucket().SetName("images")
	bucket , err := gridfs.NewBucket(d.mongoDB ,opts)
	if err != nil {
		return err
	}

	_ , err = bucket.DownloadToStreamByName(fmt.Sprintf("%v.jpg" , id) , w);
	if err != nil {
		return err
	}

	return nil
}