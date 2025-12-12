FROM golang:1.23

ARG DB_URL
ARG DB_TOKEN

LABEL maintainer="adferdev"

RUN apt-get update && apt-get install -y \
    git \
    build-essential \
    pkg-config

WORKDIR /app

COPY . .

RUN --mount=type=secret,id=DB_URL \
    --mount=type=secret,id=DB_TOKEN \
    --mount=type=secret,id=API_PROD_URL \
    echo "TURSO_DB_URL=$(cat /run/secrets/DB_URL)" > .env && \
    echo "TURSO_DB_TOKEN=$(cat /run/secrets/DB_TOKEN)" >> .env && \
    echo "API_ENVIRONMENT=production" >> .env && \
    echo "API_PROD_URL_HOST=$(cat /run/secrets/API_PROD_URL)" >> .env && \
    echo "API_CACHE_EXPIRATION=1h" >> .env && \
    echo "API_CACHE_EVICTION_INTERVAL=10m" >> .env

RUN go get -d -v ./...

RUN go install -v ./...

RUN go build -o /build

EXPOSE 8080

CMD [ "/build" ]
