deploy:
	GOOS=linux GOARCH=arm64 go build -o build/hws-rpi cmd/hws/main.go
	scp build/hws-rpi root@hkk-pi.local:/opt/hws/
	ssh root@hkk-pi.local "chmod +x /opt/hws/hws-rpi && /opt/hws/hws-rpi" | hl