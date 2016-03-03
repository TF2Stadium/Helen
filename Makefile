default: assets static

clean:
	rm assets/bindata.go assets/geoip.mmdb
	go clean
assets:
	wget "http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.mmdb.gz" -O "assets/geoip.mmdb.gz"
	gzip -d -f assets/geoip.mmdb.gz

	go-bindata -ignore="bindata\.go" \
	-pkg assets -tags bindata 	 \
	-o assets/bindata.go assets/

static: assets/geoip.mmdb assets/bindata.go
	go build -tags "netgo bindata" -ldflags "-w -linkmode external -extldflags -static" -v -o Helen

docker: 
	go build -tags "netgo bindata" -ldflags "-w -linkmode external -extldflags -static" -v -o Helen
	docker build -t tf2stadium/helen .

cover:
#	sh -ex cover.sh
