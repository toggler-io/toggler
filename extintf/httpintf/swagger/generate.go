package swagger

//go:generate rm -f ./specfs/fs.go
//go:generate esc -o ./specfs/fs.go -pkg specfs swagger.json

//go:generate rm -f ./uifs/fs.go
//go:generate esc -o ./uifs/fs.go -pkg uifs -prefix assets/swagger-ui ./assets

//go:generate swagger generate spec --work-dir .. --output swagger.json
//go:generate swagger validate swagger.json

//go:generate rm -rf ${WDP}/lib/go
//go:generate mkdir -p ${WDP}/lib/go
//go:generate swagger generate client --quiet --spec swagger.json --target ${WDP}/lib/go
