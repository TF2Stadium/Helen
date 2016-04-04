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
	CGO_ENABLED=0 go build -tags "netgo bindata" -v 						\
	-ldflags 											\
	"-X \"github.com/TF2Stadium/Helen/internal/version.GitCommit=`git rev-parse HEAD`\" 		\
	-X \"github.com/TF2Stadium/Helen/internal/version.GitBranch=`git rev-parse --abbrev-ref HEAD`\" \
	-X \"github.com/TF2Stadium/Helen/internal/version.BuildDate=`date`\" 				\
	-X \"github.com/TF2Stadium/Helen/internal/version.Hostname=`hostname`\""			\
	-o Helen

tests:
	go test -tags bindata -race -v `go list ./... | grep -v /vendor/`
cover:
#	sh -ex cover.sh
