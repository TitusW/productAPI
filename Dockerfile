FROM golang:latest

LABEL maintainer="Titus <tituswus@gmail.com>"

WORKDIR /app

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY . .

ENV PORT=8080 \
    SECRET_KEY=123 \
    MongoDb=mongodb+srv://Sagara:pwsagara@productapi.u6s11.mongodb.net/myFirstDatabase?retryWrites=true&w=majority

RUN go build

CMD ["./productAPI"]