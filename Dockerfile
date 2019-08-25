FROM golang
RUN go get github.com/toggler-io/toggler/cmd/toggler
ENTRYPOINT toggler http-server -port 8080
EXPOSE 8080