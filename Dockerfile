FROM scratch

WORKDIR /app

COPY web ./web

COPY ./bin/scheduler .

ENV TODO_PORT=7540
ENV TODO_DBFILE=scheduler.db

EXPOSE ${TODO_PORT}/tcp

CMD ["/app/scheduler"]