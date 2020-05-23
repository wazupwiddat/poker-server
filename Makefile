POKER_BIN := bin
POKER_SERVER := poker-server
POKER_SRC := cmd/main.go
STATIC_FILES = server/templates
ASSET_FILES = server/assets


all: build copy build-image

build: $(POKER_SRC)
	go build -o $(POKER_BIN)/$(POKER_SERVER) $(POKER_SRC)

copy:
	mkdir -p $(POKER_BIN)/templates
	mkdir -p $(POKER_BIN)/assets/{js,css,images}
	cp -f $(STATIC_FILES)/* $(POKER_BIN)/templates
	cp -r $(ASSET_FILES)/ $(POKER_BIN)/assets/

build-image:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 026450499422.dkr.ecr.us-east-1.amazonaws.com
	docker build -t poker-service .
	docker tag poker-service:latest 026450499422.dkr.ecr.us-east-1.amazonaws.com/poker-service:latest
	docker push 026450499422.dkr.ecr.us-east-1.amazonaws.com/poker-service:latest

run-image:
	docker run -it -p 8080:8080 -v ~/.aws:/.aws poker-service:latest