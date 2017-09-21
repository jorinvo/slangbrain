PROD := fbot.slangbrain.com
PROD_DB := $(PROD):/etc/slangbrain/slangbrain.db
PROD_BIN := /usr/local/bin/slangbrain
DEV := tunnel.slangbrain.com

STAT := stat.slangbrain.com
STAT_BIN := slangbrain-stat
STAT_FILE := /usr/local/bin/$(STAT_BIN)

VERSION := $(shell go version | cut -d' ' -f3)_commit_$(shell git log --format="%H" -n 1)

MIGRATION := $(shell ls -1 migrations | tail -n1)
MIGRATION_FILE := ./migrations/$(MIGRATION)/main.go
NUM := $(shell echo "$(MIGRATION)" | cut -d_ -f1)
BEFORE_DB := backups/$(NUM)-before.db
AFTER_DB := backups/$(NUM)-after.db

FB_TOKEN := EAAEMXBS5vNoBAB4NbuAYJp1FhDN50UNcoFRtME4phWQEGdV3ezUUkZCVS6B1Q2vQHFPc4TUZBdMTwEWjkwzfFNR2WR5cYDxjXZCZCWKUgAlZBewOGKZB1Un2gSbaBKV2L4bgn7vR5ZC83Lo7kd53WZAimctkkwRzBmzA1UYRvR0sgQZDZD
FB_SECRET := 414a6cdf4c8da5bbb281960cfcfe3eeb
VERIFY_TOKEN := SmhklHbrVi4MInnC8Fih58TBTIc$$jeXTadn%bChS

SLACK_HOOK := https://hooks.slack.com/services/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



# Run dev server locally
run:
	@go run main.go \
		-db 'dev.db' \
		-http 8080 \
		-domain $(DEV) \
		-verify $(VERIFY_TOKEN) \
		-token $(FB_TOKEN) \
		-secret $(FB_SECRET) \
		-slackhook $(SLACK_HOOK) \
		$@

.PHONY: run



# Run integration tests
test:
	@go test -v \
		-coverpkg ./api,./brain,./common,./fbot,./messenger,./payload,./slack,./translate,./user \
		./integration

.PHONY: test



# Build, and deploy latest version of Slangbrain to the live server
deploy:
	GOOS=linux go build -a -ldflags "-s -w -X main.version=$(VERSION)" -o dist/slangbrain
	scp dist/slangbrain $(PROD):/tmp/slangbrain
	@echo "switch to new binary and restart service"
	@ssh -t $(PROD) "sh -c 'sudo mv $(PROD_BIN) /tmp/slangbrain-$(shell date +%s)' && sudo mv /tmp/slangbrain $(PROD_BIN) && sudo systemctl reload slangbrain.service && sudo journalctl -fu slangbrain"

.PHONY: deploy


# Deploy slangbrain-stat binary
deploy-stat:
	GOOS=linux go build -a -ldflags "-s -w" -o dist/$(STAT_BIN) ./cmd/$(STAT_BIN)/main.go
	ssh $(STAT) mv $(STAT_FILE) /tmp/$(STAT_BIN)-$(shell date +%s)
	scp dist/$(STAT_BIN) $(STAT):$(STAT_FILE)

.PHONY: deploy-stat



# Apply most recent migration on production server
migrate:
	@mkdir -p backups

	errcheck $(MIGRATION_FILE)

	ssh -t $(PROD) sudo systemctl stop slangbrain

	scp $(PROD_DB) $(BEFORE_DB)
	cp $(BEFORE_DB) $(AFTER_DB)

	go run $(MIGRATION_FILE) $(AFTER_DB)
	@sleep 2

	scp $(AFTER_DB) $(PROD_DB)
	deploy

.PHONY: migrate



# Rollback data to before the last migration
rollback:
	ssh -t $(PROD) sudo systemctl stop slangbrain
	scp backups/$(shell echo "$(MIGRATION)" | cut -d_ -f1)-before.db $(PROD_DB)
	ssh -t $(PROD) "sh -c 'sudo systemctl restart slangbrain && sudo journalctl -fu slangbrain'"

.PHONY: rollback
