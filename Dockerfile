FROM golang:1.13-alpine
ENV CGO_ENABLED=0
WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build

FROM alpine:3.11
COPY --from=0 /go/src/app/vaultsync /vaultsync
ENTRYPOINT [ "/vaultsync" ]
