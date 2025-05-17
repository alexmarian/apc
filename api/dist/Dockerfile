FROM almalinux:8 AS builder

# Install Go
RUN dnf install -y epel-release && \
    dnf install -y golang-1.23 git make && \
    dnf clean all

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM scratch AS export-stage
COPY --from=builder /app/main /main
CMD ["/bin/bash"]