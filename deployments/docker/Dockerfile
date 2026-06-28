# syntax=docker/dockerfile:1

FROM node:20-alpine AS web
WORKDIR /src/web
RUN corepack enable
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

FROM golang:1.26-alpine AS backend
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/dist ./pkg/embed/web/dist
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-w -s -X 'kube-tide/configs.BuildMode=production'" \
    -o /out/kube-tide ./cmd/server/main.go

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=backend /out/kube-tide /app/kube-tide
COPY configs/config.yaml /app/configs/config.yaml
ENV K8S_PLATFORM_ENV=production
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/kube-tide"]
