run:
	go run cmd/main.go

bufgen:
	cd proto && rm -rf generated && buf generate