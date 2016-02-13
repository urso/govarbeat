BEATNAME=govarbeat
BEAT_DIR=github.com/urso
SYSTEM_TESTS=false
TEST_ENVIRONMENT=false
ES_BEATS=./vendor/github.com/elastic/beats
GOPACKAGES=$(shell glide novendor)
PREFIX?=.

# Path to the libbeat Makefile
-include $(ES_BEATS)/libbeat/scripts/Makefile

.PHONY: init
init:
	glide update  --no-recursive
	make update
	git init
	git add .

.PHONY: install-cfg
install-cfg:
	mkdir -p $(PREFIX)
	cp etc/govarbeat.template.json     $(PREFIX)/govarbeat.template.json
	cp etc/govarbeat.yml               $(PREFIX)/govarbeat.yml
	cp etc/govarbeat.yml               $(PREFIX)/govarbeat-linux.yml
	cp etc/govarbeat.yml               $(PREFIX)/govarbeat-binary.yml
	cp etc/govarbeat.yml               $(PREFIX)/govarbeat-darwin.yml
	cp etc/govarbeat.yml               $(PREFIX)/govarbeat-win.yml

.PHONY: update-deps
update-deps:
	glide update  --no-recursive
