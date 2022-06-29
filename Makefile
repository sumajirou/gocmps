gocmps: *.go
	go build -o gocmps *.go

test: gocmps
	./test.sh

clean:
	rm -f gocmps tmp*

.PHONY: test clean
