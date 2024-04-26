FROM golang:alpine3.19 as build-kustomize

RUN apk update && apk add make gcc git musl-dev openssh

RUN git clone --depth=1 --single-branch https://github.com/kubernetes-sigs/kustomize.git && \
    cd kustomize && \
    make kustomize

FROM golang:alpine3.19 as build

ARG VERSION
ARG GIT_COMMIT
ARG GIT_DESCRIPTION
ARG BUILD_DATE

# install build dependencies
RUN apk update && apk add git --no-cache

WORKDIR /app/src

COPY . .

# pre-fetch dependencies
RUN go mod download

RUN mkdir -p ../bin && \
    go build -ldflags="-X 'main.cliVersion=${VERSION}' -X 'main.gitCommit=${GIT_COMMIT}' -X 'main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -X 'main.gitDescription=${GIT_DESCRIPTION}'" -o ../bin/deploykit .

FROM alpine:latest

COPY --from=build /app/bin/deploykit /usr/local/bin/
COPY --from=build-kustomize /go/bin/kustomize /usr/local/bin/

RUN apk update && \
    apk add git

LABEL org.opencontainers.image.created=${BUILD_DATE} \
    org.opencontainers.image.authors="Bacchus Jackson" \
    org.opencontainers.image.url="https://github.com/gatecheckdev/deploykit" \
    org.opencontainers.image.source="https://github.com/gatecheckdev/deploykit" \
    org.opencontainers.image.version=${VERSION} \
    org.opencontainers.image.revision=${GIT_COMMIT} \
    org.opencontainers.image.licenses="Apache 2.0" \
    org.opencontainers.image.ref.name=${GIT_COMMIT} \
    org.opencontainers.image.title="GitOps DeployKit" \
    org.opencontainers.image.description="A simple utility for performing common GitOps tasks" \
    org.opencontainers.image.documentation="https://github.com/gatecheckdev/deploykit/blob/main/README.md"
