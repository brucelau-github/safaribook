build: main.go
	go build -o zeen main.go
	./zeen epub -k conf 0131429019
