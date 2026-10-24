### Build Go application ###
FROM golang:1.25-bookworm AS builder

WORKDIR /app

# Copy the local ocr module first so the go.mod replace directive
# (replace github.com/hththai/ocr => ../ocr) resolves to /ocr inside
# the builder. The build context must be the project root for this to work.
COPY ocr/ /ocr/

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ .
RUN go build -o main .

### Final Image ###
FROM debian:stable-slim

WORKDIR /app

# Install system packages and Maldet dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    bash \
    wget \
    tar \
    adduser \
    curl \
    ca-certificates \
    gosu \
    libmupdf-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Maldet, pinned to a release. Do not use master: v2.x made its library
# files unreadable to non-root users, so scans as appuser fail with exit 127.
ARG MALDET_VERSION=1.6.6.1
RUN wget https://github.com/rfxn/linux-malware-detect/archive/refs/tags/${MALDET_VERSION}.tar.gz -O maldetect.tar.gz \
    && mkdir maldetect-src \
    && tar -xzf maldetect.tar.gz -C maldetect-src --strip-components=1 \
    && cd maldetect-src && ./install.sh \
    && cd /app \
    && rm -rf maldetect-src maldetect.tar.gz

# Add Maldet to PATH
ENV PATH="/usr/local/maldetect:${PATH}"

# Enable user public scanning
RUN sed -i 's/^scan_user_access=.*/scan_user_access=1/' /usr/local/maldetect/conf.maldet \
    && grep '^scan_user_access' /usr/local/maldetect/conf.maldet

# Create non-root appuser and required data directories
RUN groupadd -g 1000 appuser \
    && useradd -u 1000 -g appuser -m appuser \
    && mkdir -p /app/app/log /app/filedata \
    && chown -R appuser:appuser /app

# Create pub paths (must be done as root)
RUN /usr/local/maldetect/maldet --mkpubpaths

# Copy Go binary and entrypoint from builder
COPY --from=builder /app/main /app/main
COPY server/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8088

ENTRYPOINT ["/entrypoint.sh"]
