package main

import (
	"log"
	"net/http"
	"os"

	"juno.api/handlers"
	"juno.api/internal"

	"github.com/rs/cors"
)

type HttpHandler func (w http.ResponseWriter , r *http.Request)

func POST(w http.ResponseWriter , r *http.Request , handler HttpHandler) HttpHandler{
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w , http.StatusText(http.StatusMethodNotAllowed) , http.StatusMethodNotAllowed)
			return
		}
		handler(w , r)
	}	
}

func GET(w http.ResponseWriter , r *http.Request , handler HttpHandler) HttpHandler{
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w , http.StatusText(http.StatusMethodNotAllowed) , http.StatusMethodNotAllowed)
			return
		}
		handler(w , r)
	}	
}

func main(){
	mux := http.NewServeMux();
	mux.HandleFunc("/" , func (w http.ResponseWriter , r *http.Request){
		w.Write([]byte("Hello World!"))
	})

	db := internal.Database{}
	db.Init()

	app := handlers.App{
		Database: db,
	}

	
	ws := &internal.WebSocket{};
	ws.Init(mux , "/feed" , app.WSFeed);

	mux.HandleFunc("/upload" , app.UploadFile); // POST : Upload a file to the database using gridFS
	mux.HandleFunc("/file", app.DownloadFile);	// GET : download file from the database using gridFS

	mux.HandleFunc("/signUp" , app.SignUp); // POST 
	mux.HandleFunc("/signIn" , app.SignIn);	// POST 
	mux.HandleFunc("/refresh" , app.Refresh);	// GET : refresh authentication token

	mux.HandleFunc("/details" , app.Details);	// GET : Get user account details


	mux.HandleFunc("/products" , app.Products); // GET get top n product recommendations 
	mux.HandleFunc("/search" , app.SearchProducts); // GET : search products database given a query
	mux.HandleFunc("/liked" , app.Liked); // GET : get all products liked by user
	mux.HandleFunc("/cart" , app.Cart); // GET : Get user's shopping cart

	
	handler := cors.New(cors.Options{
		AllowedOrigins : []string{
			"*",
		},
		AllowCredentials : true,


		Debug : false,
	}).Handler(mux)

	PORT := os.Getenv("PORT")
	log.Println("Running and serving on PORT" , PORT)
	err :=  http.ListenAndServe("0.0.0.0:" + PORT , handler)
	if err != nil {
		log.Println("failed to serve http , err =" , err)
	}	
}
