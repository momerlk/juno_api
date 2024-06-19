package handlers

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	"io/ioutil"
	"net/http"
)

func (a *App) UploadFile(w http.ResponseWriter, r *http.Request) {



    log.Println("File Upload Endpoint Hit")

    // Parse our multipart form, 10 << 20 specifies a maximum
    // upload of 10 MB files.
    r.ParseMultipartForm(10 << 20)
    // FormFile returns the first file for the given key `myFile`
    // it also returns the FileHeader so we can get the Filename,
    // the Header and the size of the file
    file, handler, err := r.FormFile("file")
    if err != nil {
        log.Println("Error Retrieving the File")
        log.Println(err)
        return
    }
    defer file.Close()
    log.Printf("Uploaded File: %+v\n", handler.Filename)
    log.Printf("File Size: %+v\n", handler.Size)
    log.Printf("MIME Header: %+v\n", handler.Header)



    // read all of the contents of our uploaded file into a
    // byte array
    fileBytes, err := ioutil.ReadAll(file)
    if err != nil {
        log.Println(err)
    }

    id := uuid.NewString()
    

    a.Database.StoreJPG(id , fileBytes)

    // return that we have successfully uploaded our file!
    json.NewEncoder(w).Encode(bson.M{"id" : id})
}

func (a *App) DownloadFile(w http.ResponseWriter , r *http.Request){
    id := r.URL.Query().Get("id")

    w.Header().Set("Content-Type", "application/octet-stream")


    a.Database.GetJPG(id , w)
}