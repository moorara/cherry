# BUILD STAGE
FROM golang:1.11-alpine as builder
RUN apk add --no-cache git
WORKDIR /workspace
COPY . .
ENV CGO_ENABLED=0
RUN git checkout "$(git tag --list | tail -n 1)" && go install && git checkout -
RUN cherry build -cross-compile=false -binary-file=cherry

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
