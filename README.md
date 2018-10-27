# Slangbrain

[Slangbrain](https://slangbrain.com/) is a Facebook Messenger Chatbot to help you studying language similar to software such as [Anki](https://apps.ankiweb.net).
To learn about building chatbots and the things to consider have a look at https://jorin.me/chatbots.pdf.
I want to open source Slangbrain to provide a real world example covering all aspects of a Go project.
I think the code base shows simple, pragmatic solutions in contrast to building a Google-scale product.
Going with simple solutions can give you a nice feeling of being in control of as many parts of your application as possible.
The Chatbot has been running without any error for more than a year now.

The setup is very customized for this use case and doesn’t scale, but it solved all my needs and is simple to think about.

_The following are some highlights of what I think was fun to work on or is solved in unconventional ways:_

It sends me weekly reports in Slack looking like this:

```
users:          61
subscriptions:   8
dbsize:        3.45mb

format:         total (avg)
---------------------------
phrases:         7825 (128)
scoretotal:      5534 (90)
studies:         4693 (76)
due studies:     1818 (29)
imports:           51 (0)
notifies:         189 (3)
zeroscore:        956 (15)
new phrases:     6051 (99)
```

The statistics are generate by a [separate binary](/cmd/slangbrain-stat/main.go) running against the latest backup file which doesn't slow down the main DB and validates that the backup is actually working.


All interactions are done through a [Makefile](/Makefile):

* run locally with ssh tunnel
* test
* deploy
* trigger backup
* run migrations
* see production logs
* update dependencies

The app has zero-downtime deploys using a [systemd socket file](/slangbrain.socket).

It can run without proxy server.
HTTPS is done via the Go package for letsencrypt: golang.org/x/crypto/acme/autocert

For editing mode it uses a [webview](/webview) with the HTML embedded in the same Go binary.

All HTTP is done using the Go [standard library](https://golang.org/pkg/net/http/) and very little dependencies are used. All dependencies are [commited with the source code](/vendor).

The app is internationalized. [Translations](/translate/lang_en.go) are simply Go structs.

There are integrations available to automate importing and exporting data and more.
Automation can be done using the [HTTP API](https://slangbrain.com/api/) or through uploading files from URL or as CSV files.

Testing is done through full [integration tests](/integration) simulating HTTP requests in the same way Facebook will actually send webhooks.

[Migrations](/migrations) are separate binaries which are run before starting the main app and discarded after.

[Slack](/slack/slack.go) is used as admin interface. Errors and statistics are reported here. When users send feedback it's directly send to a Slack channel and an admin can reply to the feedback from within there.

The business and DB logic ([brain](/brain)) is separated from Facebook Messenger specific code ([bot](/bot)). This would also allow for porting the functionality to other platforms easier.

Users are notified when it's time for them to study. Timers are [tracked](/bot/notify.go) in memory. Some work is put into taking care of details such as not sending notifications at the [night time of a user](https://github.com/jorinvo/slangbrain/blob/9dfa7ed04fca9fdeccf73fabdd45de1e65e60c03/brain/study.go#L138)

The current [mode](/brain/brain.go#L22) of each user's chat session must be tracked server-side.

No external database is used. All data is stored in a single file. [boltdb](https://github.com/coreos/bbolt) is used for storage. Data is encoded using [gob](https://golang.org/pkg/encoding/gob/). Since data is simply stored as key-value pairs and There are [more buckets](/brain/bucket/bucket.go#L8) than you would have tables in a relational DB. Aggregates such as a user's score are tracked separate at write-time. Can think about this as creating your own (dumb) indexes.

Backups are done [through HTTP](/main.go#L169) and triggered from a external script.

Users get a [weekly report](/brain/userstats.go#L41) about their progress for some gamification:

  * weekly added
  * weekly studied
  * total study score
  * global ranking compared to others

The website is also open source at https://github.com/jorinvo/slangbrain.com. It is build with the Go static site generator [Hugo](https://gohugo.io/).



## Development

Use `make` for all tasks.


## License

[MIT](/LICENSE)
