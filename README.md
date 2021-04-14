# Discord Notification

This project provides a POST webhook to send the notification to discord channel. This can be modified to make it work for any type of message or alert for your team on Discord.

### Usage

Git clone this repo or use go lang to get the latest version of the code. Once it's cloned locally cd into discordnotification and start downloading dependencies as show below. 

``` bash
  $ go get https://github.com/runicstudios/discordnotification
  $ cd discordnotification
  $ go get . && go build -o app
```

Note: You should have go version 1.16.x installed to run this package

Once building is done a binary executable with name <b>app</b> will be created in the same directory, use that to initialize the http server

```bash
   $ ./app --discord "<callback webhook url for discord>"
```
