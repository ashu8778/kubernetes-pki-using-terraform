FROM golang:1.22.1 AS stage1
WORKDIR /api
COPY main.go go.mod /api
RUN CGO_ENABLED=0 go build -o go-api

FROM alpine
WORKDIR /api
COPY --from=stage1 /api/go-api /api
ENTRYPOINT ["/api/go-api"]