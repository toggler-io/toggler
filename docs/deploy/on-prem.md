# Deploying to On-Premises

to deploy to your own server, you need to build the binary first.

* clone the project
* cd to the project root directory
* `go build -o toggler cmd/toggler/main.go`
* use your favorite CD tooling to deliver the binary to the server
* To see how to run the application check [Procfile](Procfile) for example.
