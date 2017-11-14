FROM busybox:1.24.2

WORKDIR /app

RUN addgroup -g 10001 app && \
    adduser -G app -u 10001 -D -h /app -s /sbin/nologin app

COPY version.json /app/version.json
COPY main /app/main
RUN touch /etc/policies.yaml  # No policy by default.

USER app

ENV GIN_MODE release
ENV POLICIES /etc/policies.yaml
ENV PORT 8000

ENTRYPOINT ["/app/main"]
