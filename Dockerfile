# ==========================================
# STAGE 1: Build the custom Go binary
# ==========================================
FROM golang:1.25.5-alpine AS builder

LABEL description="Go Builder for Custom Free5GC NEF"

# Install Git and build tools needed for Go modules
RUN apk add --no-cache gcc musl-dev git

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your custom source code into the container
COPY . .

# Compile the Go code into a binary named 'nef'
# (CGO_ENABLED=0 ensures it runs smoothly on Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -o build/bin/nef cmd/main.go

# ==========================================
# STAGE 2: Run the binary in a clean environment
# ==========================================
FROM alpine:3.13

LABEL description="free5GC NEF service" version="Stage 3"

ENV F5GC_MODULE nef
ARG DEBUG_TOOLS

# Install debug tools ~ 100MB (if DEBUG_TOOLS is set to true)
RUN if [ "$DEBUG_TOOLS" = "true" ] ; then apk add -U vim strace net-tools curl netcat-openbsd ; fi

# Setup the free5gc user and permissions
Run addgroup -S free5gc && adduser -S free5gc
Run mkdir -p /free5gc && chown -R free5gc:free5gc /free5gc
USER free5gc

# Set working dir
WORKDIR /free5gc
RUN mkdir -p config/ cert/ log/

# Copy the compiled binary from STAGE 1
COPY --from=builder /app/build/bin/${F5GC_MODULE} ./

# Exposed ports
EXPOSE 8000
