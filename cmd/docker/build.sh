name="quark-signin"
GOOS=linux GOARCH=amd64 go build -v -ldflags="-w -s" -o ./bin/$name