#!/usr/bin/env bash

function download_assets {
    echo "Downloading assets"
    curl -o "assets/geoip.mmdb.gz" \
	 "http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.mmdb.gz"
    gzip -d -f assets/geoip.mmdb.gz
}

if [ ! -f assets/geoip.mmdb ]; then
    download_assets
fi

if [ "$1" = "production" ]; then
    echo "Compiling assets"
    go-bindata -nomemcopy -ignore="bindata\.go" \
	       -pkg assets 		        \
	       -o assets/bindata.go assets/
else
    echo "Compiling assets (debug)"
    go-bindata -debug -ignore="bindata\.go" \
	       -pkg assets    		    \
	       -o assets/bindata.go assets/

fi

