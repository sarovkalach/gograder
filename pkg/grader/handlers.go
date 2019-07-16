package grader

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// func main() {
//     http.HandleFunc("/", HelloServer)
//     http.ListenAndServe(":8080", nil)
// }

func (g *Grader) ReceiveTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only Post Method"))
		return
	}
	log.Println("Income message")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	task := &Task{}
	if err := json.Unmarshal(body, task); err != nil {
		http.Error(w, "Error in unmarshalling JSON", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
