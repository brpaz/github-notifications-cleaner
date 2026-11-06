# =========================================
# Build stage
# =========================================
FROM --platform=$BUILDPLATFORM golang:1.27.3-alpine3.24@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb as build

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
FROM alpine:3.24@sha256:3f4d6b32e26703c6082839234f43769d5046b3ac7c910033911d1e71d8cd630a

COPY --from=build /out/github-notifications-cleaner /bin

ENTRYPOINT ["/bin/github-notifications-cleaner"]
