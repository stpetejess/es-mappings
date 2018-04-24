PKG=github.com/stpetejess/es-mappings
GOPATH:=$(PWD)/.root:$(GOPATH)
export GOPATH

all: test

.root/src/$(PKG):
	mkdir -p $@
	for i in $$PWD/* ; do ln -s $$i $@/`basename $$i` ; done

root: .root/src/$(PKG)

clean:
	rm -rf .root
	rm -rf tests/*_es-mappings.go

build:
	go build -i -o .root/bin/es-mappings $(PKG)/es-mappings

generate: root build
	.root/bin/es-mappings -all -output_filename=.root/src/$(PKG)/tests/feed_item_mapping_actual.json .root/src/$(PKG)/tests/data.go

test: generate root
	go test \
		$(PKG)/tests \
		$(PKG)/gen


.PHONY: root clean generate test build
