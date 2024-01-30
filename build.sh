env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o bin/ImageProcessor main.go
git add .
git commit -m "Build: bin"
git push origin production