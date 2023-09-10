# Use an official Ubuntu image as the base
FROM ubuntu:latest

# Install required packages
RUN apt-get update && apt-get install -y \
    sudo \
    python3 \
    python3-pip \
    zip \
    unzip \
    curl

# Install s3cmd using pip
RUN pip3 install s3cmd

# Fetch the latest release from the coreweave/gutenberg-epub repository and download it
RUN mkdir -p /app && \
    curl https://github.com/coreweave/gutenberg-epub/releases/download/release/gutenberg-epub-converter -o /app/gutenberg-epub && \
    chmod +x /app/gutenberg-epub

# Copy the entrypoint.sh script into the container
COPY entrypoint.sh /app/entrypoint.sh

# Make the entrypoint.sh script executable
RUN chmod +x /app/entrypoint.sh

# Set the entry point to the shell script
ENTRYPOINT ["/app/entrypoint.sh"]