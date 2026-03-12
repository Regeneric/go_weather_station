deploy:
	GOOS=linux GOARCH=arm64 go build -o build/wbs-rpi apps/wbs/main.go
	scp build/wbs-rpi root@hkk-pi.local:/opt/wbs/
	ssh root@hkk-pi.local "chmod +x /opt/wbs/wbs-rpi && cd /opt/wbs/ && ./wbs-rpi" | hl --follow
