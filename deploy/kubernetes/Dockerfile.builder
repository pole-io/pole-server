FROM golang:1.25-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends nodejs npm ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY . .

CMD ["sh", "-c", "sleep infinity"]
