FROM golang

COPY . /app

WORKDIR /app/cmd/main/

RUN go get -d -v
RUN go build main.go

CMD ["./main"]
#

