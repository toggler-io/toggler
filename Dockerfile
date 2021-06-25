FROM golang:1.16.5-alpine as build
EXPOSE 8080

WORKDIR /src/
COPY . .

ENV WDP="/src" \
	GO111MODULE=on
ENV PATH="${PATH}:${WDP}/bin:${WDP}/.tools"

RUN bin/provision
RUN CGO_ENABLED=0 go build -o /bin/toggler cmd/toggler/main.go

FROM scratch
COPY --from=build /bin/toggler /bin/toggler

ENTRYPOINT ["/bin/toggler"]
CMD ["http-server", "-port", "8080"]
