package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func jsonPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "OK"}`))
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	ParseExif(body)
	if err != nil {
		fmt.Println(err)
	}
	err = ioutil.WriteFile("test.json", body, os.ModeAppend)
	if err != nil {
		fmt.Println(err)
	}
}
func filePost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	// max buffer size, not max file size
	r.ParseMultipartForm(32 << 20)

	file, handler, err := r.FormFile("mediafile")

	if err != nil {
		fmt.Println("Error retrieving the file")
		fmt.Println(err)
		return
	}
	defer file.Close()

	jsonData, _, err := r.FormFile("metadata")
	if err != nil {
		fmt.Println("error getting json")
		fmt.Println(err)
	}

	jsonString, _ := ioutil.ReadAll(jsonData)

	fmt.Println(string(jsonString))

	defer jsonData.Close()

	fmt.Printf("File: %+v\n", handler.Filename)
	fmt.Printf("Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create a temp file within a temp-images dir
	tempFile, err := ioutil.TempFile("temp-images", "upload-*.png")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tempFile.Close()

	// read all bytes

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = tempFile.Write(fileBytes)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// at some point I'll persist this
var filesSeen map[string]int

func dupeCheck(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sha256 := params["sha256"]

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if filesSeen != nil {
		seen := filesSeen[sha256]
		if seen < 1 {
			w.Write([]byte(`{"message": "NOT_DUPE"}`))
		} else {
			w.Write([]byte(`{"message": "IS_DUPE"}`))
		}
		return
	}

	filesSeen = make(map[string]int)
	w.Write([]byte(`{"message": "NOT_DUPE"}`))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/mediafile", jsonPost).Methods(http.MethodPost)
	r.HandleFunc("/file", filePost).Methods(http.MethodPost)
	r.HandleFunc("/dupe/{sha256}", dupeCheck).Methods(http.MethodGet)
	log.Fatal(http.ListenAndServe(":8081", r))

}
