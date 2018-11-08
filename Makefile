run:
	go run ./bot

build: 
	go build -o bin/bot github.com/buckley-w-david/anibot/bot

install:
	go install github.com/buckley-w-david/anibot/bot

deploy:
	git push heroku master
	heroku container:release bot

setup:
	heroku create anibot --manifest
