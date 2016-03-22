FROM alpine

RUN apk add --update  --repository http://mirror.nl.leaseweb.net/alpine/main ca-certificates

ENV HELEN_GEOIP=true
ENV HELEN_SERVER_ADDR=0.0.0.0:80
ENV HELEN_PROFILER_ADDR=0.0.0.0:81

ADD Helen /bin/helen
ADD views /views

ENTRYPOINT ["/bin/helen"]
EXPOSE 80
EXPOSE 81