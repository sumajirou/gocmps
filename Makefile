gocmps: gocmps.go
	go build gocmps.go

test: gocmps
	./test.sh

clean:
	rm -f gocmps tmp*

.PHONY: test clean