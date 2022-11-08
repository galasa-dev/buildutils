all: bin/galasabld-linux-amd64 \
	bin/galasabld-windows-amd64 \
	bin/galasabld-darwin-amd64 \
	bin/galasabld-linux-s390x \
	bin/galasabld-darwin-arm64

bin/galasabld-linux-amd64 : ./Makefile ./cmd/galasabld/main.go ./pkg/cmd/*.go ./pkg/galasayaml/*.go ./pkg/githubjson/*.go 
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/galasabld-linux-amd64 ./cmd/galasabld

bin/galasabld-windows-amd64 : ./Makefile ./cmd/galasabld/main.go ./pkg/cmd/*.go ./pkg/galasayaml/*.go ./pkg/githubjson/*.go 
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/galasabld-windows-amd64 ./cmd/galasabld

bin/galasabld-darwin-amd64 : ./Makefile ./cmd/galasabld/main.go ./pkg/cmd/*.go ./pkg/galasayaml/*.go ./pkg/githubjson/*.go 
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/galasabld-darwin-amd64 ./cmd/galasabld

bin/galasabld-darwin-arm64 : ./Makefile ./cmd/galasabld/main.go ./pkg/cmd/*.go ./pkg/galasayaml/*.go ./pkg/githubjson/*.go 
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/galasabld-darwin-arm64 ./cmd/galasabld

bin/galasabld-linux-s390x : ./Makefile ./cmd/galasabld/main.go ./pkg/cmd/*.go ./pkg/galasayaml/*.go ./pkg/githubjson/*.go 
	CGO_ENABLED=0 GOOS=linux GOARCH=s390x go build -o bin/galasabld-linux-s390x ./cmd/galasabld


clean:
	rm -rf bin
