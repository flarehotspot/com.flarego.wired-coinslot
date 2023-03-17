default:
	go build -tags=dev -trimpath -buildmode=plugin -o plugin.so

prod:
	go build -buildmode=plugin -o plugin.so -tags=prod ./main.go
