# Start from a minimal base image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the compiled binary from your local system to the container
COPY main .

# Expose the port your binary listens on (e.g., 8000)
EXPOSE 8000

# Set the entrypoint to run the binary
ENTRYPOINT ["./main"]
