FROM jgsqware/go-glide

WORKDIR /app

ENV SRC_DIR=/go/src/iot-stats/
# Source code:
ADD . $SRC_DIR
# Build it:
ADD config.json /app/
ADD build /app/
ADD server.pem /app/
ADD server.key /app/
RUN cd $SRC_DIR; glide install; go build -o iot-stats; cp iot-stats /app/

ENTRYPOINT ["./iot-stats"]
EXPOSE 3000