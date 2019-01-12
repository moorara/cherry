# BUILD STAGE
FROM golang:1.11-alpine as builder
RUN apk add --no-cache git
WORKDIR /workspace
COPY . .
RUN CGO_ENABLED=0 ./scripts/build.sh --main main.go --binary cherry

# FINAL STAGE
FROM golang:1.11-alpine
RUN apk add --no-cache ca-certificates git
RUN apk add --no-cache ruby && \
    gem install rdoc --no-document && \
    gem install github_changelog_generator
COPY --from=builder /workspace/cherry /usr/local/bin/
RUN chown -R nobody:nogroup /usr/local/bin/cherry
USER nobody
ENTRYPOINT [ "cherry" ]
