FROM golang:1.18-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 go build -o /prometheus-gitlab-license-exporter

FROM scratch

COPY --from=build /prometheus-gitlab-license-exporter .

CMD ["/prometheus-gitlab-license-exporter"]
