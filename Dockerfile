FROM golang:alpine AS builder

# Add all the source code (except what's ignored
# under `.dockerignore`) to the build context.
ADD ./ /go/src/github.com/karan/TakeoverBot/

RUN set -ex && \
  cd /go/src/github.com/karan/TakeoverBot && \
  go build -o main && \
  mv ./main /usr/bin/main

FROM busybox

# Retrieve the binary from the previous stage
COPY --from=builder /usr/bin/main /usr/local/bin/main

# Set the binary as the entrypoint of the container
ENTRYPOINT [ "main" ]
