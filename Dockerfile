# Barebones Dockerfile for testing httpGO
FROM alpine

COPY httpGO /usr/local/bin/httpGO

RUN chmod +x /usr/local/bin/httpGO

RUN addgroup -S myawesomegroup
RUN adduser -S myawesomeuser -G myawesomegroup

USER myawesomeuser

# can we make this configurable?
EXPOSE 8000

CMD /usr/local/bin/httpGO