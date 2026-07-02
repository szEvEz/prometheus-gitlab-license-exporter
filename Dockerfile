FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /prometheus-gitlab-license-exporter

FROM scratch

COPY --from=build /prometheus-gitlab-license-exporter .

CMD ["/prometheus-gitlab-license-exporter"]
