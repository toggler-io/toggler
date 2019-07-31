package httpintf

import (
	"net/http"
)

//go:generate swagger generate spec -o swagger.json
//go:generate swagger validate swagger.json
//go:generate esc -private -o ./swagger-assets.go -pkg httpintf swagger.json
//go:generate swagger generate client --quiet --spec swagger.json --target ${WDP}/lib/go

func HandleSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(`Content-Type`, `application/json`)
	w.WriteHeader(200)
	_, _ = w.Write(_escFSMustByte(false, `/swagger.json`))
}
