FROM golang:alpine as builder

COPY . /app

WORKDIR /app/cmd/main

RUN CGO_ENABLED=0

RUN go build main.go


FROM alpine

COPY --from=builder /app/cmd/main/main .

CMD ["./main"]
#

