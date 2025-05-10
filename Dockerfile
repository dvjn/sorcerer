FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

ENV GOPATH=/app/.cache
COPY go.mod go.sum ./
RUN --mount=type=cache,target=${GOPATH} \
    go mod download

ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
COPY . .
RUN --mount=type=cache,target=${GOPATH} \
    go build -o sorcerer ./cmd/sorcerer

FROM scratch

WORKDIR /app
COPY --from=build /app/sorcerer .

ENTRYPOINT ["./sorcerer"]
