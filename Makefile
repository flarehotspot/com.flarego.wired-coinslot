plugin: export CGO_ENABLED=0

plugin:
	go build -buildmode=plugin -o plugin.so ./main.go
