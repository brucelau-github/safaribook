build: main.go
	go build -o zeen main.go
	./zeen epub -k conf 0596007736
