FROM golang:1.19-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o idgen-server

FROM alpine:3.17
WORKDIR /app
COPY --from=0 /app/idgen-server .
CMD /app/idgen-server