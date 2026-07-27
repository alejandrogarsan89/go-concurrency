# Build a tiny static image for the demo CLI using a multi-stage build.

# ---- build stage ----
FROM golang:1.22-alpine AS build
WORKDIR /src

# Cache module downloads first.
COPY go.mod go.sum ./
RUN go mod download

# Build the statically-linked binary.
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/demo ./cmd/demo

# ---- runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/demo /demo
ENTRYPOINT ["/demo"]
CMD ["--help"]
