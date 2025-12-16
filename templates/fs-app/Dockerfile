FROM node:lts-slim AS node-build

WORKDIR /go/src/app
COPY . .

RUN npm install
RUN npm run build
RUN npm run tailwind

FROM golang:1.25-alpine AS go-build

WORKDIR /go/src/app
COPY . .

COPY --from=node-build /go/src/app/internal/server/assets/js/app.js ./internal/server/assets/js/app.js
COPY --from=node-build /go/src/app/internal/server/assets/css/styles.css ./internal/server/assets/css/styles.css
RUN go mod download
RUN go tool templ generate

RUN CGO_ENABLED=0 go build -o /go/bin/app cmd/server/main.go
FROM gcr.io/distroless/static-debian12:latest AS go-app

COPY --from=go-build /go/bin/app /
EXPOSE 8080
CMD ["/app"]
