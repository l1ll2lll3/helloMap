package main

import (
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	createEmployeeHanlder := http.HandlerFunc(createEmployee)
	http.Handle("/photo", createEmployeeHanlder)
	http.ListenAndServe(":8080", nil)
}

func createEmployee(w http.ResponseWriter, request *http.Request) {
	d, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	tmpfile, err := os.Create("./" + "photo.png")
	defer tmpfile.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	tmpfile.Write(d)

	w.WriteHeader(200)
	return
}
