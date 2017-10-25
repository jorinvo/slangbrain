prod := fbot.slangbrain.com
prod_path := /usr/local/bin/slangbrain
dev := tunnel.slangbrain.com

stat := stat.slangbrain.com
stat_bin := slangbrain-stat
stat_path := /usr/local/bin/$(stat_bin)

version := $(shell go version | cut -d' ' -f3)_commit_$(shell git log --format="%H" -n 1)

MIGRATION ?= $(shell ls -1 migrations | tail -n1)
migration_file := ./migrations/$(MIGRATION)/main.go

# Credentials for local development
fb_token := EAAEMXBS5vNoBAB4NbuAYJp1FhDN50UNcoFRtME4phWQEGdV3ezUUkZCVS6B1Q2vQHFPc4TUZBdMTwEWjkwzfFNR2WR5cYDxjXZCZCWKUgAlZBewOGKZB1Un2gSbaBKV2L4bgn7vR5ZC83Lo7kd53WZAimctkkwRzBmzA1UYRvR0sgQZDZD
fb_secret := 414a6cdf4c8da5bbb281960cfcfe3eeb
verify_token := SmhklHbrVi4MInnC8Fih58TBTIc3jeXTadn1bChS
slack_hook := https://hooks.slack.com/services/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



# Run dev server locally and make it available publicly via ssh tunnel
run:
	-@$(MAKE) -j tunnel server



# Run dev server locally
server:
	-@go run main.go \
		-db 'dev.db' \
		-http 8080 \
		-domain $(dev) \
		-verify $(verify_token) \
		-token $(fb_token) \
		-secret $(fb_secret) \
		-slackhook $(slack_hook) \
		-nosetup



# Start a tunnel to the dev server
tunnel:
	@echo "tunneling port 8080 to $(dev)"
	-@ssh $(dev) -NR 8080:localhost:8080



# Run integration tests
test:
	go test ./integration



# Run tests verbose and output coverage
test-cover:
	@go test -v \
		-coverpkg ./api,./bot,./brain,./payload,./scope,./slack,./translate,./webview \
		./integration



# Lint code
lint:
	golint ./... | grep -v vendor



# Build, and deploy latest version of Slangbrain to the live server
deploy: lint test
	GOOS=linux go build -a -ldflags "-s -w -X main.version=$(version)" -o dist/slangbrain
	scp dist/slangbrain $(prod):/tmp/slangbrain
	@echo "switch to new binary and reload service"
	-@ssh -t $(prod) "sh -c 'sudo mv $(prod_path) /tmp/slangbrain-$(shell date +%s)' && sudo mv /tmp/slangbrain $(prod_path) && sudo systemctl reload slangbrain.service && sudo journalctl -fu slangbrain"



# Deploy slangbrain-stat binary
deploy-stat: lint test
	GOOS=linux go build -a -ldflags "-s -w" -o dist/$(stat_bin) ./cmd/$(stat_bin)/main.go
	scp dist/$(stat_bin) $(stat):/tmp/$(stat_bin)
	@echo "switch to new binary"
	@ssh -t $(stat) "sh -c 'sudo mv $(stat_path) /tmp/$(stat_bin)-$(shell date +%s) && sudo mv /tmp/$(stat_bin) $(stat_path)'"



# Trigger backup job
backup:
	ssh -t $(stat) "sh -c 'sudo systemctl kill -s ALRM slangbrain-backup'"



# Apply most recent migration on production server
# Make sure to `make backup` upfront.
# You can run old migrations like using `MIGRATION=006_bucket_scoretotals make migrate`.
# Migrations take a path to a database file as first and only arg.
migrate:
	@echo "using migration $(MIGRATION)"
	GOOS=linux go build -a -ldflags "-s -w" -o dist/$(MIGRATION) $(migration_file)
	scp dist/$(MIGRATION) $(prod):/tmp/

	@echo "activate migration and reload service"
	-@ssh -t $(prod) "sh -c 'sudo chown slangbrain:root /tmp/$(MIGRATION) && sudo mv /tmp/$(MIGRATION) /etc/slangbrain/migrations/ && sudo systemctl reload slangbrain && sudo journalctl -fu slangbrain'"



# Show logs of production server
logs:
	-ssh -t $(prod) sudo journalctl -fu slangbrain



# Update vendored dependencies using dep
update-deps:
	go get github.com/golang/dep
	dep ensure
	dep ensure -update
	dep prune

# Remove created artifacts and dev DB
clean:
	rm dev.db
	rm dist/slangbrain dist/slangbrain-stat



.PHONY: run test deploy deploy-stat backup migrate update-deps clean
