
test:
	go clean -testcache
	cd core/base && go test -v ./...
	cd user/assets && go test -v ./...
	# cd user/orders && go test -v ./...