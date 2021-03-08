FROM golang:1.14-alpine as build
EXPOSE 8080

WORKDIR /src/
COPY . .
RUN set -e; \
	. .envrc.build; \
	bin/provision; \
	CGO_ENABLED=0 go build -o /bin/toggler cmd/toggler/main.go

FROM scratch
COPY --from=build /bin/toggler /bin/toggler

ENTRYPOINT ["/bin/toggler"]
CMD ["http-server", "-port", "8080"]
