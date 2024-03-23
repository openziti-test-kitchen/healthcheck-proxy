ARG BUILDER_BASE=golang:1.22-alpine3.18
FROM ${BUILDER_BASE} AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go program
RUN go build -o zerotrust-healthcheck .

# Final stage
FROM alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the built executable from the builder stage
COPY --from=builder /app/zerotrust-healthcheck .

# Command to run the executable
CMD ["./zerotrust-healthcheck"]