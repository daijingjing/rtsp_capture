
.PHONY: all
all:
	@$(MAKE) clean capture

.PHONY: clean
clean:
	rm -rf build

.PHONY: deps
deps:
	go mod download -x
	mkdir -p build

.PHONY: capture
capture: deps
	go build -ldflags "-s -w" -o build/capture_rtsp ./cmd

.PHONY: deploy
deploy:
	scp build/capture_rtsp ${DEPLOY_IP}:/usr/bin/capture_rtsp

%: DEPLOY_IP:=192.168.254.254
%: export CGO_ENABLED:=1
%: export GOOS=linux
%: export GOARCH=arm64
#%: export GOARM=5