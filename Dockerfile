FROM golang:1.19

# WORKDIR /usr/src/avito-user-balance
ENV GOPATH=/

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
# COPY go.mod ./
COPY ./ ./
RUN go mod download && go mod tidy && go mod verify

# RUN ls /usr/local/go/src/avito-user-balance/
# RUN echo $GOROOT
# RUN go mod download
RUN go build -o avito-app ./cmd/main.go

CMD ["./avito-app"]
