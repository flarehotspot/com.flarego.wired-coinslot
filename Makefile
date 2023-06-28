plugin: export CGO_ENABLED=1

plugin:
	go build -buildmode=plugin -o plugin.so ./main.go
