build: 
	go build -o bin/bot github.com/buckley-w-david/anibot/bot
	go build -o bin/hooks github.com/buckley-w-david/anibot/hooks

install:
	go install github.com/buckley-w-david/anibot/bot
	go install github.com/buckley-w-david/anibot/hooks

deploy:
	git push heroku master
	heroku container:release web bot


setup:
	heroku create anibot --manifest
