FROM golang:1.21-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /bin/url-shortener ./cmd/server

FROM alpine:3.19
COPY --from=build /bin/url-shortener /usr/local/bin/url-shortener
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/url-shortener"]
