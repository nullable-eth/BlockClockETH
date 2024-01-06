FROM golang:1.19.4-alpine3.17
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o BlockClockETH .
RUN cp /app/BlockClockETH /
RUN cp /app/config.json /
RUN rm -rf /app
WORKDIR /
CMD ["/BlockClockETH"]