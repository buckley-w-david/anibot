build:
	go build -o bin/bot github.com/buckley-w-david/anibot/bot

install:
	go install github.com/buckley-w-david/anibot/bot

run:
	go run github.com/buckley-w-david/anibot/bot

deploy:
	git push heroku master

setup:
	heroku create anibot --manifest
