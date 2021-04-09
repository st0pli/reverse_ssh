ADDR=localhost:2200


debug: 
	mkdir -p bin
	go build -o bin ./...

release:
	mkdir -p bin
	go build -ldflags="-s -w" -o bin ./...

run: 
	./bin/client --reconnect $(ADDR)
	./bin/client --reconnect $(ADDR)
	./bin/client --reconnect $(ADDR)
	cd bin; ./server $(ADDR)
	

	