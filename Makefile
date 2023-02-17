default:
	go build -tags=dev -trimpath -buildmode=plugin -o plugin.so
