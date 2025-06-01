FROM golang:1.21.7-alpine3.19 AS build
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/
COPY . .
RUN go mod download
RUN go build -o /go/bin/eventstore ./cmd/eventstore/main.go

FROM scratch AS release
LABEL org.opencontainers.image.vendor="Openyard ES³"
LABEL org.opencontainers.image.authors="tomask@openyard.de"
COPY --from=build /go/bin/eventstore /bin/eventstore
ENTRYPOINT ["/bin/eventstore"]