FROM golang:alpine as builder
COPY ./udsTest.go /cndp/udsTest.go
WORKDIR /cndp
RUN go build -o udsTest ./udsTest.go

FROM alpine:3.14
COPY --from=builder /cndp/udsTest /bin/udsTest
RUN apk add -U iproute2
