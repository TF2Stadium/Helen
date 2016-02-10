FROM alpine

ADD helen /helen
ADD views /views
#ADD static /static
ADD lobbySettingsData.json /lobbySettingsData.json
ADD geoip.mmdb /geoip.mmdb
ENV HELEN_GEOIP_DB=geoip.mmdb

ENTRYPOINT /helen