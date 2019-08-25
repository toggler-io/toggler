package swagger

//go:generate esc -o ./specfs/fs.go -pkg specfs swagger.json
//go:generate esc -o ./uifs/fs.go -pkg uifs -prefix assets/swagger-ui ./assets

//go:generate swagger generate spec -b .. -o swagger.json
//go:generate swagger validate swagger.json
//go:generate swagger generate client --quiet --spec swagger.json --target ${WDP}/lib/go
