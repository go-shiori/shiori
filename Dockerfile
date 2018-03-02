FROM golang:1.10.0-alpine3.7

WORKDIR /go/src/shiori
COPY . .

# Install git and gcc
# Get dependencies
# Install dependencies
# Create shiori.db as a file, so in case that the file
# is mounted with -v, a folder will not be created
RUN apk --no-cache add git build-base \
&& go get -d -v ./... \
&& go install -v ./... \
&& touch shiori.db

CMD ["shiori", "serve"]
