default: assets static

clean:
	rm assets/bindata.go assets/geoip.mmdb
	go clean
assets:
	curl -o "assets/geoip.mmdb.gz" "http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.mmdb.gz"
	gzip -d -f assets/geoip.mmdb.gz

	go-bindata -ignore="bindata\.go" \
	-pkg assets -tags bindata 	 \
	-o assets/bindata.go assets/

static: assets/geoip.mmdb assets/bindata.go
	CGO_ENABLED=0 go build -tags "netgo bindata" -v -o Helen

tests:
	go test -v -race -tags bindata ./...
cover:
#	sh -ex cover.sh
