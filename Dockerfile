############################
# STEP 1 build executable binary
############################
# golang alpine 1.15.5
FROM golang@sha256:4d8abd16b03209b30b48f69a2e10347aacf7ce65d8f9f685e8c3e20a512234d9 as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create appuser
ENV USER=appuser
ENV UID=10001
ENV GOARCH=amd64
ENV BLOGPATH=/blog

# See https://stackoverflow.com/a/55757473/12429735
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"
WORKDIR $GOPATH/src/github.com/sksmith/blog-server
COPY . .
RUN chmod 777 $GOPATH/src/github.com/sksmith/blog-server/run.sh

# Fetch dependencies.
RUN go get -d -v

# Build the binary
RUN VER=$(git describe --tag);TIM=$(date +'%Y-%m-%d_%T');SHA1=$(git rev-parse HEAD); \
        CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build \
      -ldflags='-w -s -extldflags "-static" -X main.AppVersion='$VER' -X main.BuildTime='$TIM' -X main.Sha1Version='$SHA1 -a \
      -o /go/bin/blog-server .

CMD $GOPATH/src/github.com/sksmith/blog-server/run.sh