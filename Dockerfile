FROM alpine

RUN adduser -h /home/faasnats -D faasnats faasnats

COPY ./faas-nats /home/faasnats/
RUN chmod +x /home/faasnats/faas-nats

USER faasnats

ENTRYPOINT ["/home/faasnats/faas-nats"]
