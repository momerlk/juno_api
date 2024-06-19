package handlers 

import (
	"log"

	"pakmaweshi.api/internal"
	"net/http"
)

type App struct {
	Database 			internal.Database
}

func (a *App) ServerError(w http.ResponseWriter , reqName string , err error){
	log.Printf("%v : Internal Error encountered : %v" , reqName ,  err)	
	http.Error(w , http.StatusText(http.StatusInternalServerError) , http.StatusInternalServerError)
}

func (a *App) ClientError(w http.ResponseWriter , code int) {
	http.Error(w , http.StatusText(code) , code)
}