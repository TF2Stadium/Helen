default: geoip docker

geoip:
	wget "http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.mmdb.gz" -O "geoip.mmdb.gz"
	gzip -d geoip.mmdb.gz

docker: geoip.mmdb
	cp -R ../views/ .
	cp ../lobbySettingsData.json .
	go build -ldflags "-linkmode external -extldflags -static" -v  -o helen ../
	docker build -t tf2stadium/helen .

clean:
	rm -rf views
	rm -rf lobbySettingsData.json
	rm -rf helen
