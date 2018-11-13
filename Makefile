run:
	go run ./cmd/bot

build: 
	go build -o bin/bot ./cmd/bot

install:
	go install ./cmd/bot

deploy:
	git push heroku master
	heroku container:release bot

setup:
	heroku create anibot --manifest
