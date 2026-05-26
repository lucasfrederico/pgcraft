# Multi-stage build: bin compilado vai num scratch image (~15MB total).
# Multi-arch via buildx (linux/amd64, linux/arm64).
#
# Build local:
#   docker build -t pgcraft .
# Run:
#   docker run -it --network host pgcraft "postgres://localhost/mydb"

FROM golang:1.25-alpine AS builder
WORKDIR /src

# cache de modules separado dos sources — só re-baixa quando go.mod muda
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# -trimpath: builds reprodutíveis
# -ldflags "-s -w": tira symbol/DWARF info → binário menor (~30% redução)
# CGO_ENABLED=0: estático, roda em scratch
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" \
    -o /out/pgcraft ./cmd/pgcraft

# Runtime: alpine pra TTY/terminfo bonito (scratch mata cores Bubble Tea)
FROM alpine:3.20
RUN apk add --no-cache ncurses ca-certificates
COPY --from=builder /out/pgcraft /usr/local/bin/pgcraft
ENTRYPOINT ["pgcraft"]
