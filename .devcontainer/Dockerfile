FROM golang:1.26

WORKDIR /app

# Install system packages and configure locale
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    curl \
    gh \
    git \
    gnupg \
    jq \
    locales \
    unzip \
    && rm -rf /var/lib/apt/lists/* \
    && sed -i -E 's/# (ja_JP.UTF-8)/\1/' /etc/locale.gen \
    && locale-gen

ENV LANG=ja_JP.UTF-8

# Install Go tools
RUN go install golang.org/x/tools/cmd/goimports@latest && \
    go install golang.org/x/tools/gopls@latest && \
    go install honnef.co/go/tools/cmd/staticcheck@latest && \
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 && \
    go install golang.org/x/vuln/cmd/govulncheck@latest

# Install Terraform
RUN curl -fsSL https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(. /etc/os-release && echo $VERSION_CODENAME) main" > /etc/apt/sources.list.d/hashicorp.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends terraform \
    && rm -rf /var/lib/apt/lists/*

# Install AWS CLI v2 (architecture-aware, no official apt package available)
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then AWS_ARCH="x86_64"; \
    elif [ "$ARCH" = "aarch64" ]; then AWS_ARCH="aarch64"; \
    else echo "Unsupported architecture: $ARCH" && exit 1; fi && \
    curl -fsSL "https://awscli.amazonaws.com/awscli-exe-linux-${AWS_ARCH}.zip" -o /tmp/awscliv2.zip \
    && unzip /tmp/awscliv2.zip -d /tmp \
    && /tmp/aws/install \
    && rm -rf /tmp/awscliv2.zip /tmp/aws
