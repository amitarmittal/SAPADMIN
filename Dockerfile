# Building the binary of the App
FROM  REPOSITORY/golang:1.17 AS build

# `SP` should be replaced with your project name
WORKDIR /go/src/sp

# Copy all the Code and stuff to compile everything
COPY . .

# Downloads all the dependencies in advance (could be left out, but it's more clear this way)
RUN go mod download

# Builds the application as a staticly linked one, to allow it to run on alpine
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -o app .


# Moving the binary to the 'final Image' to make it smaller
FROM REPOSITORY/alpine:latest

WORKDIR /app

# `Gap` should be replaced here as well
COPY --from=build /go/src/sp/app .

# Exposes port 3000 because our program listens on that port
EXPOSE 3000

CMD ["./app"]
