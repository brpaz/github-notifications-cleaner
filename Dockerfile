# =========================================
# Build stage
# =========================================
FROM --platform=$BUILDPLATFORM golang:1.26.5-alpine3.24@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 as build

ARG TARGETOS
ARG TARGETARCH
ARG BUILD_DATE
ARG GIT_COMMIT
ARG GIT_VERSION

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify && go mod tidy

RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build \
    -ldflags="-s -w \
    -X github.com/brpaz/github-notifications-cleaner/cmd/version.BuildDate=${BUILD_DATE} \
    -X github.com/brpaz/github-notifications-cleaner/cmd/version.Version=${GIT_VERSION} \
    -X github.com/brpaz/github-notifications-cleaner/cmd/version.GitCommit.=${GIT_COMMIT} \
    -extldflags '-static'" -a \
    -o /out/github-notifications-cleaner ./main.go

# ====================================
# Production stage
# ====================================
FROM alpine:3.24@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

COPY --from=build /out/github-notifications-cleaner /bin

ENTRYPOINT ["/bin/github-notifications-cleaner"]
