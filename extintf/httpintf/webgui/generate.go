//go:generate esc -o ./assets/fs.go -ignore fs.go -pkg assets -prefix assets ./assets
//go:generate esc -o ./views/fs.go  -ignore fs.go -pkg views  -prefix views  ./views
package webgui
