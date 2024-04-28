FROM --platform=${TARGETPLATFORM:-linux/amd64} scratch
# FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:latest

# these parameters are required
# example: --build-arg SERVICE_NAME=go-service-template --build-arg GOOS=linux --build-arg GOARCH=arm64
ARG SERVICE_NAME
ARG BUILD_DATE
ARG BUILD_VERSION
ARG GOOS
ARG GOARCH

LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=${BUILD_DATE}
LABEL org.label-schema.name=${SERVICE_NAME}
LABEL org.label-schema.version=$BUILD_VERSION

# make available the service name in the container
ENV SERVICE_NAME=${SERVICE_NAME}

WORKDIR /app
ENV PATH="/app:${PATH}"

COPY "dist/${SERVICE_NAME}-${GOOS}-${GOARCH}" /app/microservice

ENTRYPOINT ["/app/microservice"]
