#!/usr/bin/env bash

function download_assets {
    if [ -z "${MAXMIND_API_KEY}" ] {
       echo "Must set MAXMIND_API_KEY to download required GeoIP files"
       exit 1
    }

    echo "Downloading assets"
    curl -o "assets/geoip.mmdb.gz" \
	 "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=${MAXMIND_API_KEY}&suffix=tar.gz\"
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
