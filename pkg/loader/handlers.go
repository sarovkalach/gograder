package loader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/gorilla/mux"
)

type Server struct {
	Router   *mux.Router
	Uploader *FileLoader
}

func NewServer() *Server {
	s := &Server{Router: mux.NewRouter(), Uploader: NewFileLoader()}
	s.initRoutes()
	return s
}

func (s *Server) initRoutes() {
	s.Router.HandleFunc("/", s.uploadFile).Methods("GET")
	s.Router.HandleFunc("/upload", s.uploadSuccess).Methods("POST")
}

func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("../../web/templates/upload.html"))
	t.Execute(w, nil)
}

func (s *Server) uploadSuccess(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Huge file"))
	}
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()

	f, err := os.OpenFile("/tmp/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	io.Copy(f, file)

	meta := make(map[string]string)
	meta["filename"] = handler.Filename
	if err := s.Uploader.saveUserTask(meta); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Internal server Error"))
	}
	t := template.Must(template.ParseFiles("../../web/templates/success.html"))
	t.Execute(w, nil)
}

func (s *Server) Run() {
	log.Println("Server Start")
	log.Fatal(http.ListenAndServe(":8080", s.Router))
}
