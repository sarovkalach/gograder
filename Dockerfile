# Start from golang:1.11-alpine base image
FROM golang:last

# Add Maintainer Info
LABEL maintainer="Kalachov Alex <akalachov@mail.ru>"

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/sarovkalach/gograder

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Download all the dependencies
# https://stackoverflow.com/questions/28031603/what-do-three-dots-mean-in-go-command-line-invocations
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Run the executable
CMD ["go-docker-compose"]