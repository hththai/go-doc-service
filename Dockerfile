### Build Go application ###
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

### Final Image ###
FROM debian:stable-slim

WORKDIR /app

# Install CA certificates to enable outgoing HTTPS requests.
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Install system packages and Maldet dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    bash \
    wget \
    tar \
    adduser \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Maldet
# Install Maldet
RUN wget http://www.rfxn.com/downloads/maldetect-current.tar.gz \
    && tar -xzf maldetect-current.tar.gz \
    && cd maldetect-* && ./install.sh \
    && cd /app \
    && rm -rf maldetect-* maldetect-current.tar.gz

# Add Maldet to PATH
ENV PATH="/usr/local/maldetect:${PATH}"

# Enable user public scanning
RUN sed -i 's/^scan_user_access=.*/scan_user_access=1/' /usr/local/maldetect/conf.maldet \
    && grep '^scan_user_access' /usr/local/maldetect/conf.maldet

# Create non-root appuser
RUN groupadd -g 1000 appuser \
    && useradd -u 1000 -g appuser -m appuser \
    && chown -R appuser:appuser /app

# Create pub paths (must be done as root)
RUN /usr/local/maldetect/maldet --mkpubpaths

# Copy Go binary from builder
COPY --from=builder /app/main /app/main

USER appuser

EXPOSE 8088

CMD ["/app/main"]
