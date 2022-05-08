version = $(file < VERSION)

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/ddns-aws

install:
	install -m 755 ./bin/ddns-aws /usr/local/bin
	install -m 644 ./.ddns-aws.yaml /etc/

service:
	install -m 644 ./ddns-aws.service /etc/systemd/system/

docker:
	docker build -t psihoza/ddns-aws:latest .
	docker tag psihoza/ddns-aws:latest psihoza/ddns-aws:${version}

uninstall:
	rm /usr/local/bin/ddns-aws
	rm /etc/.ddns-aws.yaml
	rm /etc/systemd/system/ddns-aws.service

clean:
	go clean
	rm bin/*
