FROM node:18-alpine AS frontend-builder

WORKDIR /app
COPY web/frontend/package*.json ./
RUN npm install
COPY web/frontend ./
RUN npm run build

FROM golang:1.22-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
COPY --from=frontend-builder /app/dist ./web/frontend/dist

RUN go build -o taskflow ./cmd/taskflow

FROM alpine:latest

RUN apk --no-cache add ca-certificates bash

WORKDIR /app

COPY --from=backend-builder /app/taskflow /app/taskflow

RUN mkdir -p /app/data

ENV PORT=8080
ENV DB_PATH=/app/data/taskflow.db
ENV JWT_SECRET=""
ENV LOG_LEVEL=info

EXPOSE 8080

CMD ["./taskflow"]
