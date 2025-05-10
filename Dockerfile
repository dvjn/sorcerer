FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o sorcerer ./cmd/sorcerer

FROM scratch

WORKDIR /app
COPY --from=build /app/sorcerer .

ENTRYPOINT ["./sorcerer"]
