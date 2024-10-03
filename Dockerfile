# Usage: docker build -t vreedb .
# Usage: docker run -p 8080:8080 vreedb

# Start from the latest golang base image
FROM golang:latest

# Add Maintainer Info
LABEL maintainer="spookieoli"

# Install gcc and other necessary tools
RUN apt-get update && apt-get install -y gcc

# Set the Current Working Directory inside the Docker container
WORKDIR /

# Create the collections folder
RUN mkdir -p /collections

# Copy the source from the current directory to the Working Directory inside the Docker container
COPY . .

# Disable GOLANG telemetry
CMD ["go", "telemetry", "off"]

# Build the Go app
RUN go build -o VreeDB .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./VreeDB", "-avx256=True"]