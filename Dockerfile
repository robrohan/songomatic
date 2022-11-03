FROM golang:alpine as builder
RUN apk --no-cache add gcc g++ make ca-certificates git
WORKDIR /go/src/github.com/robrohan/go-web-template
COPY . .
RUN make build

FROM golang:alpine
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/robrohan/go-web-template/build/ ./
RUN ls -alFh
CMD ["./server"]
