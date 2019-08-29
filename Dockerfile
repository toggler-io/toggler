FROM golang:1.12
RUN go get github.com/toggler-io/toggler/cmd/toggler
ENTRYPOINT ["toggler"]
CMD ["http-server", "-port", "8080"]
EXPOSE 8080