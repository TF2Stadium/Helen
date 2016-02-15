default: assets static

clean:
	rm assets/bindata.go assets/geoip.mmdb
	go clean
assets:
	wget "http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.mmdb.gz" -O "assets/geoip.mmdb.gz"
	gzip -d assets/geoip.mmdb.gz

	go-bindata -ignore="bindata\.go" \
	-pkg assets -tags bindata 	 \
	-o assets/bindata.go assets/

static: assets/geoip.mmdb assets/bindata.go
	go build -tags bindata -ldflags "-w -linkmode external -extldflags -static" -v -o Helen

docker: 
	go build -tags bindata -ldflags "-w -linkmode external -extldflags -static" -v -o Helen
	docker build -t tf2stadium/helen .

cover:
	# go get github.com/axw/gocov/gocov
	# go get gopkg.in/matm/v1/gocov-html

	# gocov test ./models/ | gocov-html > models.html

	# git clone git@github.com:TF2Stadium/coverage.git
	# cp models.html ./coverage/
	# cd coverage
	# git config --global user.email "this@is.bot"
	# git config --global user.name "circleci deploy"
	# cp index_template index.html
	# printf "$(date -u) \n</body>" >> index.html
	# git add models.html index.html
	# git commit -m "Update coverage" && git push -f
