

monitor.zip: monitor
	zip $@ $<

monitor: lambda.go
	GOOS=linux GOARCH=amd64 go build -o $@ $<

clean:
	rm -f monitor.zip monitor
