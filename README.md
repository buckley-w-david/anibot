# anibot

Search for anime and manga without leaving the comfort of your discord server!

## What does it do?
`anibot` listens for two kinds of messages in any discord server it is added to:

### Inline Requests
An inline request is any bit of text between a set of braces ({}) for anime or a set of inequality signs (<>) for manga.

Examples:  
`Have you guys heard of {Hunter x Hunter}?`  
or  
`<Berserk> is way better than the anime would lead you to believe.`

### Bot commands
A bot command is a message prefixed with `!anibot `. With bot commands you can get more specific than with the inline requests, looking up media based on title, ID (from anilist, where all the data is pulled from), studio, or staff.

```bash
TODO: Section on bot commands
```

### Response

The bot will respond with something that looks like...
![Hunter x Hunter anime preview](assets/hunterxhunter.png)
![Berserk manga preview](assets/berserk.png)

### What's with the reactions?

I'm so glad you asked!

Since discord is a little behind the times when it comes to methods of interacting with a bot, those of us interested in a more user friendly experience have had to do some improvising. Those reactions are ways to request additional information about the preview that was just sent.

In the above screenshot, reacting to the message with "👉" will result in the bot pulling additional pieces of media that the Director has worked on, while "👈" will request the same for the Original Creator.

This feature is still a work in progress, but expect equvilant studio requests soon!

To prevent spam, each button will only work once. After is has been pressed, and the info put into chat, using that reaction again won't do anything.

*Note*: As of now, if the bot goes offline, all previously existing buttons will unfortunatly stop functioning, even after the bot is brought back up. In practice this isn't a huge deal as the bot isn't expected to go down very often (if ever), and old buttons no longer working isn't such a big deal anyway.

## How do I use it?

1. Setup a discord application through the [Discord Developer Portal](https://discordapp.com/developers/applications/).
1. Create a bot user for that application.
1. Copy the bot token.
1. Clone this repository.
1. run `make build` inside the cloned directory (This will require you have `go` installed on your system).
1. run `bin/bot -t "<Your Bot Token from Step 3>"`

The bot should now be running, add it to a server to see it go.

To add the bot to your server, go to https://discordapp.com/oauth2/authorize?&client_id=<Your application client ID\>&scope=bot&permissions=19456.  
You'll need the client ID from the application you setup in step 1

## Why would I want this?

 ¯\_(ツ)_/¯