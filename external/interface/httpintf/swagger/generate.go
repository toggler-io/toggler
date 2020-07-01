package swagger

//= Generate documentation
//go:generate rm -f swagger.json
//go:generate swagger generate spec --work-dir ../httpapi --output swagger.json
//go:generate swagger validate swagger.json

//= Generate go client
//go:generate rm -rf lib
//go:generate mkdir -p lib
//go:generate swagger generate client --quiet --spec swagger.json --target lib

//= Embed generated documentation
//go:generate rm -f ./specfs/fs.go
//go:generate esc -o ./specfs/fs.go -pkg specfs swagger.json
//go:generate rm -f ./uifs/fs.go
//go:generate esc -o ./uifs/fs.go -pkg uifs -prefix assets/swagger-ui ./assets
//
