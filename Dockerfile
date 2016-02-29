FROM alpine

RUN apk add --update ca-certificates

ADD Helen /bin/helen
ADD views /views

ENV HELEN_GEOIP=true
ENV HELEN_SERVER_ADDR=0.0.0.0:80
ENV HELEN_PROFILER_ADDR=0.0.0.0:81

ENTRYPOINT ["/bin/helen"]
EXPOSE 80
EXPOSE 81