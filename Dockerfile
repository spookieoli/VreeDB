# Start from the latest golang base image
FROM golang:latest

# Add Maintainer Info
LABEL maintainer="spookieoli"

# Install gcc and other necessary tools
RUN apt-get update && apt-get install -y gcc

# Set the Current Working Directory inside the Docker container
WORKDIR /

# Set the default environment to DEV
ENV ENV DEV

# Create the collections folder
RUN mkdir -p /collections

# Copy the source from the current directory to the Working Directory inside the Docker container
COPY . .

# Compile the AVX check C code
RUN cd /avx && gcc -c -o avx_check.o avx_check.c && ar rcs libavx_check.a avx_check.o

# Build the Go app
RUN go build -o VreeDB .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./VreeDB"]