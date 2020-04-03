#!/usr/bin/env bash

function download_assets {
    if [ -z "${MAXMIND_API_KEY}" ]; then
       echo "Must set MAXMIND_API_KEY to download required GeoIP files"
       exit 1
    fi

    echo "Downloading assets"
    pushd assets
    curl -o "geoip.tar.gz" "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&license_key=${MAXMIND_API_KEY}&suffix=tar.gz"
    tar xzvf geoip.tar.gz --wildcards --no-anchored '*.mmdb' --strip-components=1
    mv GeoLite2-Country.mmdb geoip.mmdb
    rm geoip.tar.gz
    popd
}

if [ ! -f assets/geoip.mmdb ]; then
    download_assets
fi

if [ "$1" = "production" ]; then
    echo "Compiling assets"
    go-bindata -nomemcopy -ignore="bindata\.go" \
	       -pkg assets 		        \
	       -o assets/bindata.go assets/{geoip.mmdb,lobbySettingsData.json}
else
    echo "Compiling assets (debug)"
    go-bindata -debug -ignore="bindata\.go" \
	       -pkg assets    		    \
	       -o assets/bindata.go assets/{geoip.mmdb,lobbySettingsData.json}

fi
