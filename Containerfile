FROM --platform=${TARGETPLATFORM:-linux/amd64} scratch
# FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:latest

# these parameters are required
# example: --build-arg SERVICE_NAME=go-rest-api-service-template --build-arg GOOS=linux --build-arg GOARCH=arm64
ARG SERVICE_NAME
ARG BUILD_DATE
ARG BUILD_VERSION
ARG DESCRIPTION
ARG REPO_URL
ARG GOOS
ARG GOARCH

# https://github.com/opencontainers/image-spec/blob/main/annotations.md
LABEL org.opencontainers.image.created=${BUILD_DATE}
LABEL org.opencontainers.image.title=${SERVICE_NAME}
LABEL org.opencontainers.image.version=$BUILD_VERSION
LABEL org.opencontainers.image.description=${DESCRIPTION}
LABEL org.opencontainers.image.source=${REPO_URL}

# make available the service name in the container
ENV SERVICE_NAME=${SERVICE_NAME}

WORKDIR /app
ENV PATH="/app:${PATH}"

COPY "dist/${SERVICE_NAME}-${GOOS}-${GOARCH}" /app/microservice

ENTRYPOINT ["/app/microservice"]
