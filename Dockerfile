# Barebones Dockerfile for testing httpGO
FROM alpine

#RUN groupadd -g 3333 appuser && \
#    useradd -r -u 3333 -g appuser appuser

RUN addgroup -g 3333 appuser && \
    adduser -r -u 3333 -g appuser appuser

USER appuser

COPY httpGO /usr/local/bin/httpGO

RUN chmod +x /usr/local/bin/httpGO

ENTRYPOINT /usr/local/bin/httpGO

# can we make this configurable?
EXPOSE 8000