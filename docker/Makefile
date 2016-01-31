docker:
	cp -R ../views/ .
	cp ../lobbySettingsData.json .
	go build -ldflags "-linkmode external -extldflags -static" -v  -o helen ../
	docker build -t tf2stadium/helen .

clean:
	rm -rf views
	rm -rf lobbySettingsData.json
	rm -rf helen
