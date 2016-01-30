docker:
	go build -ldflags "-linkmode external -extldflags -static" -v -o helen
	docker build -t tf2stadium/helen .
