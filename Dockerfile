FROM golang

ADD . /go/src/github.com/adamluzsi/toggler
RUN go install /go/src/github.com/adamluzsi/toggler/cmd/toggler
# ENV DATABASE_URL required
# ENV CACHE_URL optional
ENTRYPOINT /go/bin/toggler -cmd http-server -port 8080
EXPOSE 8080