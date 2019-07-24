# Start from golang:1.11-alpine base image
FROM golang:1.12.0-alpine3.9

# Add Maintainer Info
LABEL maintainer="Kalachov Alex <akalachov@mail.ru>"

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/sarovkalach/gograder

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

ENV GO111MODULE=on

RUN apk update && apk add git && go get gopkg.in/natefinch/lumberjack.v2
# Download all the dependencies
# https://stackoverflow.com/questions/28031603/what-do-three-dots-mean-in-go-command-line-invocations
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

# This container exposes port 8080 to the outside world
EXPOSE 8080


RUN go build -o cmd/loader/loader cmd/loader/main.go
ENTRYPOINT ["./cmd/loader/loader"]
# Run the executable
# CMD ["go run main.go"]
# ENTRYPOINT []