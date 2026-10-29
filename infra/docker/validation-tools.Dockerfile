# syntax=docker/dockerfile:1
# Validation tools image (contractcheck, httpprobe, logcheck; identitysetup since G004; gotestreport since G064): test tooling
# only, never linked into core. Build context: infra/validation/tools/. Targets: check, runtime.
FROM golang:1.27.3-alpine AS build
ENV CGO_ENABLED=0 GOTOOLCHAIN=local GOFLAGS=-trimpath
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -o /out/ ./cmd/contractcheck ./cmd/httpprobe ./cmd/logcheck ./cmd/identitysetup ./cmd/gotestreport

FROM build AS check
# Formatting, vet and unit tests; runs with --network none from the module cache above.
ENV GOPROXY=off
CMD ["sh", "-c", "test -z \"$(gofmt -l .)\" || { gofmt -l .; exit 1; }; go vet ./... && go test ./..."]

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=build /out/contractcheck /out/httpprobe /out/logcheck /out/identitysetup /out/gotestreport /usr/local/bin/
USER nonroot:nonroot
