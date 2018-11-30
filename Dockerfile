FROM golang:1.11
# RUN go get -u github.com/golang/dep/cmd/dep
WORKDIR /go/src/github.com/wish/warmer/
COPY . /go/src/github.com/wish/warmer/
# RUN dep ensure
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .




FROM alpine:3.7
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/wish/warmer/warmer .
CMD /root/warmer
