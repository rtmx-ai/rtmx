# GoReleaser copies the pre-built binary into the image.
# For standalone builds: docker build --build-arg BINARY=rtmx .
FROM scratch

COPY rtmx /usr/bin/rtmx

ENTRYPOINT ["rtmx"]
