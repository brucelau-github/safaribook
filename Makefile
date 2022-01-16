build: main.go
	go build -o zeen main.go
	./zeen epub -k conf 9780134686097
