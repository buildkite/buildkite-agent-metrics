

monitor.zip: monitor
	zip $@ $<

monitor: lambda.go
#	go get .
	GOOS=linux GOARCH=amd64 go build -o $@ $<

clean:
	terraform destroy
	rm -f init.done deploy.done monitor.zip monitor

init.done:
	terraform init
	touch $@

deploy.done: init.done agent-metrics.tf monitor.zip
	terraform apply
	touch $@
