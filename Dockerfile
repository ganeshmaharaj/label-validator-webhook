FROM golang:1.14 as builder

WORKDIR /go/src/validate-labeling

COPY . ./
RUN go mod download && CGO_ENABLED=0 go build -o /go/bin/validate-labeling

FROM alpine:3.7
COPY --from=builder /go/bin/validate-labeling /validate-labeling
ENTRYPOINT ["/validate-labeling"]
