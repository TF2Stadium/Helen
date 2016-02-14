FROM alpine

ADD Helen /helen
ADD views /views
ENV HELEN_GEOIP=true

ENTRYPOINT /helen