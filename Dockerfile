# Barebones Dockerfile for testing httpGO
FROM alpine

RUN addgroup -S myawesomegroup
RUN adduser -S myawesomeuser -G myawesomegroup

USER myawesomeuser

COPY httpGO /usr/local/bin/httpGO

RUN chmod +x /usr/local/bin/httpGO

ENTRYPOINT /usr/local/bin/httpGO

# can we make this configurable?
EXPOSE 8000