FROM alpine:3.22

RUN apk add --no-cache ca-certificates curl tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo Asia/Shanghai >/etc/timezone

COPY pole-server /app/pole-server
COPY deploy/conf /app/conf
COPY deploy/kubernetes/entrypoint.sh /app/entrypoint.sh

RUN chmod 0755 /app/pole-server /app/entrypoint.sh \
    && mkdir -p /app/logs/runtime /app/.pole_data/observability/otel-events /work

WORKDIR /app

EXPOSE 8080 8090 8091 8093 8761 8848 8890 15010

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["start", "-c", "/work/pole-server.yaml", "--mode", "all"]
