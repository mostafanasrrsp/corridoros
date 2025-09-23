# Multi-stage build for CorridorOS corrd daemon
FROM rust:1.75-slim as builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy Cargo files
COPY daemon/corrd/Cargo.toml daemon/corrd/Cargo.lock ./

# Copy source code
COPY daemon/corrd/src ./src

# Build the application
RUN cargo build --release

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -r corridoros && useradd -r -g corridoros corridoros

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/target/release/corrd /usr/local/bin/corrd

# Create necessary directories
RUN mkdir -p /var/lib/corridoros /etc/corridoros && \
    chown -R corridoros:corridoros /var/lib/corridoros /etc/corridoros

# Switch to non-root user
USER corridoros

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["corrd"]
