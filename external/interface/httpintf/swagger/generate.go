package swagger

// Generate httpapi documentation
//go:generate rm -f api.json
//go:generate swagger generate spec --work-dir ../httpapi --output api.json
//go:generate swagger validate api.json
//
//go:generate rm -rf ${WDP}/lib/go
//go:generate mkdir -p ${WDP}/lib/go
//go:generate swagger generate client --quiet --spec api.json --target ${WDP}/lib/go
//

// Embed generated documentation
//go:generate rm -f ./specfs/fs.go
//go:generate esc -o ./specfs/fs.go -pkg specfs api.json

//go:generate rm -f ./uifs/fs.go
//go:generate esc -o ./uifs/fs.go -pkg uifs -prefix assets/swagger-ui ./assets
