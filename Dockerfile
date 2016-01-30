FROM scratch

ADD helen /helen
ADD views /views
#ADD static /static
ADD lobbySettingsData.json /lobbySettingsData.json

CMD ["/helen"]