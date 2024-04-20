# Usage: docker build -t vreedb .
# Usage: docker run -p 8080:8080 vreedb

# Start from the latest golang base image
FROM golang:latest

# Add Maintainer Info
LABEL maintainer="spookieoli"

# Set the Current Working Directory inside the Docker container
WORKDIR /

# Set the default environment to DEV
ENV ENV DEV

# create the collections folder
RUN mkdir -p /collections

# Copy the source from the current directory to the Working Directory inside the Docker container
COPY . .

# Build the Go app
RUN go build -o VreeDB .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./VreeDB"]