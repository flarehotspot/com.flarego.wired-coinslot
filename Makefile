# prod: export CGO_ENABLED=0
# prod: export GOOS=linux
# prod: export GOARCH=amd64

# default:
	# go build -tags=dev -trimpath -buildmode=plugin -o plugin.so

plugin:
	go build -buildmode=plugin -o plugin.so ./main.go
