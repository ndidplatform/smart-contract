package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type GetRequestParam struct {
	RequestId string `json:"requestId"`
}

// our main function
func main() {
	router := mux.NewRouter()
	router.HandleFunc("/callback", Callback).Methods("POST")
	log.Fatal(http.ListenAndServe(":3001", router))
}

func Callback(w http.ResponseWriter, r *http.Request) {

	var param GetRequestParam
	if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	fmt.Println(param.RequestId)

}

// package main

// import (
// 	"fmt"
// 	"os"
// )

// func getEnv(key, defaultValue string) string {
// 	value, exists := os.LookupEnv(key)
// 	if !exists {
// 		value = defaultValue
// 	}
// 	return value
// }

// func main() {
// 	conn := getEnv("RP_CALLBACK_URI", "")
// 	fmt.Println(conn)
// }
