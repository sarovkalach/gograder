package loader

import (
	"net/http"

	"github.com/gorilla/mux"
)

var uploadFormTmpl = []byte(`
<html>
	<body>
	<form action="/upload" method="post" enctype="multipart/form-data">
		Image: <input type="file" name="my_file">
		<input type="submit" value="Upload">
	</form>
	</body>
</html>
`)

type Server struct {
	Router   *mux.Router
	Uploader *FileLoader
}

func NewServer() *Server {
	s := &Server{Router: mux.NewRouter(), Uploader: NewFileLoader()}
	s.initRoutes()
	return s
}

func (a *Server) initRoutes() {
	a.Router.HandleFunc("/", a.uploadFile)
}

func (a *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	w.Write(uploadFormTmpl)
}
