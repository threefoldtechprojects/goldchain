all: install

daemonpkgs = ./cmd/goldchaind
clientpkgs = ./cmd/goldchainc
pkgs = $(daemonpkgs) $(clientpkgs)

version = $(shell git describe --abbrev=0 || echo 'v0.1')
commit = $(shell git rev-parse --short HEAD)
ifeq ($(commit), $(shell (git rev-list -n 1 $(version) | cut -c1-7) || echo 'false' ))
	fullversion = $(version)
	fullversionpath = \/releases\/tag\/$(version)
else
	fullversion = $(version)-$(commit)
	fullversionpath = \/tree\/$(commit)
endif

configpkg = github.com/nbh-digital/goldchain/pkg/config
ldflagsversion = -X $(configpkg).rawVersion=$(fullversion)

stdoutput = $(GOPATH)/bin
daemonbin = $(stdoutput)/goldchaind
clientbin = $(stdoutput)/goldchainc

test: fmt vet

# fmt calls go fmt on all packages.
fmt:
	gofmt -s -l -w $(pkgs)

generate:
	rivinecg generate blockchain --explorer=plainjs
	dep ensure -v -update

# vet calls go vet on all packages.
# NOTE: go vet requires packages to be built in order to obtain type info.
vet: install-std
	go vet $(pkgs)

# installs developer binaries.
install:
	go build -race -tags='dev debug profile' -ldflags '$(ldflagsversion)' -o $(daemonbin) $(daemonpkgs)
	go build -race -tags='dev debug profile' -ldflags '$(ldflagsversion)' -o $(clientbin) $(clientpkgs)

# installs std (release) binaries with profiling enabled on http on port 10501
install-profile-std:
	go build -tags='profile' -ldflags '$(ldflagsversion)' -o $(daemonbin) $(daemonpkgs)
	go build -ldflags '$(ldflagsversion)' -o $(clientbin) $(clientpkgs)

# installs std (release) binaries
install-std:
	go build -ldflags '$(ldflagsversion)' -o $(daemonbin) $(daemonpkgs)
	go build -ldflags '$(ldflagsversion)' -o $(clientbin) $(clientpkgs)

embed-explorer-version:
	$(eval TEMPDIR = $(shell mktemp -d))
	cp -r ./frontend $(TEMPDIR)
	sed -i 's/version=0/version=$(fullversion)/g' $(TEMPDIR)/frontend/explorer/public/*.html
	sed -i 's/version=null/version=\"$(fullversion)\"/g' $(TEMPDIR)/frontend/explorer/public/js/footer.js
	sed -i 's/versionpath=null/versionpath=\"$(fullversionpath)\"/g' $(TEMPDIR)/frontend/explorer/public/js/footer.js

explorer: release-dir embed-explorer-version
	tar -C $(TEMPDIR)/frontend -czvf release/explorer-latest.tar.gz explorer

release-explorer: get_hub_jwt explorer
	# Upload explorer flist
	curl -b "active-user=goldchain; caddyoauth=$(HUB_JWT)" -F file=@./release/explorer-latest.tar.gz "https://hub.grid.tf/api/flist/me/upload"
	# Merge with caddy
	curl -b "active-user=goldchain; caddyoauth=$(HUB_JWT)" -X POST --data "[\"tf-official-apps/caddy.flist\", \"goldchain/explorer-latest.flist\"]" "https://hub.grid.tf/api/flist/me/merge/caddy-explorer-latest.flist"

# create an flist and upload it to the hub
release-flist: archive get_hub_jwt
	# Upload flist
	curl -b "active-user=goldchain; caddyoauth=$(HUB_JWT)" -F file=@./release/goldchain-latest.tar.gz "https://hub.grid.tf/api/flist/me/upload"
	# Merge with ubuntu
	curl -b "active-user=goldchain; caddyoauth=$(HUB_JWT)" -X POST --data "[\"tf-bootable/ubuntu:16.04.flist\", \"goldchain/goldchain-latest.flist\"]" "https://hub.grid.tf/api/flist/me/merge/ubuntu-16.04-goldchain-latest.flist"

# create release archives: linux, mac and flist
archive: release-dir
	./release.sh archive

release-dir:
	[ -d release ] || mkdir release

get_hub_jwt: check-HUB_APP_ID check-HUB_APP_SECRET
	$(eval HUB_JWT = $(shell curl -X POST "https://itsyou.online/v1/oauth/access_token?response_type=id_token&grant_type=client_credentials&client_id=$(HUB_APP_ID)&client_secret=$(HUB_APP_SECRET)&scope=user:memberof:goldchain"))

check-%:
	@ if [ "${${*}}" = "" ]; then \
		echo "Required env var $* not present"; \
		exit 1; \
	fi

ineffassign:
	ineffassign $(pkgs)

.PHONY: all test fmt generate vet install install-profile-std install-std embed-explorer-version explorer release-explorer release-flist archive release-dir get_hub_jwt check-%
