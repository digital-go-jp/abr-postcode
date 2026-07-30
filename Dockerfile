# syntax=docker/dockerfile:1

FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum VERSION ./
RUN go mod download

COPY main.go ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/

ARG TARGETARCH
ARG COMMIT=none
RUN VERSION=$(cat VERSION) && \
    CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -ldflags "-X abr-postcode/internal/version.Version=${VERSION} -X abr-postcode/internal/version.Commit=${COMMIT}" -o abrp .

# Runtime stage for local development
FROM debian:bookworm-slim AS dev

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/abrp /app/abrp

RUN mkdir -p /app/data

EXPOSE 8080

ENV GIN_MODE=release \
    PORT=8080 \
    ABRP_DATA_DIR=/app/data/

ENTRYPOINT ["/app/abrp"]
CMD ["--help"]

# Deployment stage. The ECS task commands sync CSVs with the AWS CLI, so it
# ships only in the image that runs them. Kept last so a build without
# --target produces the deployable image.
FROM dev AS aws

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    unzip \
    && rm -rf /var/lib/apt/lists/*

RUN ARCH=$(dpkg --print-architecture) && \
    if [ "$ARCH" = "amd64" ]; then AWS_ARCH="x86_64"; \
    elif [ "$ARCH" = "arm64" ]; then AWS_ARCH="aarch64"; \
    else echo "Unsupported architecture: $ARCH" && exit 1; fi && \
    curl -fsSL "https://awscli.amazonaws.com/awscli-exe-linux-${AWS_ARCH}.zip" -o /tmp/awscliv2.zip && \
    unzip -q /tmp/awscliv2.zip -d /tmp && \
    /tmp/aws/install && \
    rm -rf /tmp/awscliv2.zip /tmp/aws
