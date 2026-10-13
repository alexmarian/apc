FROM almalinux/8-base AS builder

WORKDIR /app
RUN dnf install -y golang gcc
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o main .

FROM scratch AS export-stage
COPY --from=builder /app/main /main