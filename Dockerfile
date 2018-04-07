FROM alpine

RUN adduser -h /home/pipe -D pipe pipe

COPY ./pipe /home/pipe/
RUN chmod +x /home/pipe/pipe

USER pipe

ENTRYPOINT ["/home/pipe/pipe"]
