FROM golang:1.19.3 as builder
ENV GOPATH=/go/
USER root

WORKDIR /spi-oauth

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download
COPY static/callback_success.html static/callback_success.html
COPY static/callback_error.html static/callback_error.html
COPY static/redirect_notice.html static/redirect_notice.html

# Copy the go sources
COPY main.go main.go
COPY controllers/ controllers/

# build service
# Note that we're not running the tests here. Our integration tests depend on a running cluster which would not be
# available in the docker build.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o spi-oauth main.go

# Compose the final image
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.6-941

# Install the 'shadow-utils' which contains `adduser` and `groupadd` binaries
RUN microdnf install shadow-utils \
	&& groupadd --gid 65532 nonroot \
	&& adduser \
		--no-create-home \
		--no-user-group \
		--uid 65532 \
		--gid 65532 \
		nonroot

COPY --from=builder /spi-oauth/spi-oauth /spi-oauth
COPY --from=builder /spi-oauth/static/callback_success.html /static/callback_success.html
COPY --from=builder /spi-oauth/static/callback_error.html /static/callback_error.html
COPY --from=builder /spi-oauth/static/redirect_notice.html /static/redirect_notice.html

WORKDIR /
USER 65532:65532

ENTRYPOINT ["/spi-oauth"]

