default:
	go build -tags=dev -trimpath -buildmode=plugin -o plugin.so

prod:
	go build -o plugin.so -tags=prod ./main.go
