FROM golang:alpine

COPY . /app

WORKDIR /app/cmd/main

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go get -d -v
RUN go build main.go


FROM alpine

COPY --from=0 /app/cmd/main/main .

CMD ["./main"]
#

