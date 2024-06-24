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

	mux.HandleFunc("/upload" , app.UploadFile);
	mux.HandleFunc("/file", app.DownloadFile)
	mux.HandleFunc("/signUp" , app.SignUp)
	mux.HandleFunc("/signIn" , app.SignIn)

	mux.HandleFunc("/details" , app.Details)


	mux.HandleFunc("/products" , app.Products);
	mux.HandleFunc("/liked" , app.Liked);

	
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
