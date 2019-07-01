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

# vet calls go vet on all packages.
# NOTE: go vet requires packages to be built in order to obtain type info.
vet: install-std
	go vet $(pkgs)

# installs developer binaries.
install:
	go build -race -tags='dev debug profile' -ldflags '$(ldflagsversion)' -o $(daemonbin) $(daemonpkgs)
	go build -race -tags='dev debug profile' -ldflags '$(ldflagsversion)' -o $(clientbin) $(clientpkgs)

# installs std (release) binaries
install-std:
	go build -ldflags '$(ldflagsversion)' -o $(daemonbin) $(daemonpkgs)
	go build -ldflags '$(ldflagsversion)' -o $(clientbin) $(clientpkgs)

# create an flist and upload it to the hub
release-flist: archive get_hub_jwt
	curl -b "active-user=goldchain; caddyoauth=$(HUB_JWT)" -F file=@./release/goldchain-latest.tar.gz "https://hub.grid.tf/api/flist/me/upload"

# create release archives: linux, mac and flist
archive: release-dir
	./release.sh

release-dir:
	[ -d release ] || mkdir release

get_hub_jwt: check-HUB_APP_ID check-HUB_APP_SECRET
	$(eval HUB_JWT = $(shell curl -X POST "https://itsyou.online/v1/oauth/access_token?response_type=id_token&grant_type=client_credentials&client_id=$(HUB_APP_ID)&client_secret=$(HUB_APP_SECRET)&scope=user:memberof:goldchain"))

check-%:
	@ if [ "${${*}}" = "" ]; then \
		echo "Required env var $* not present"; \
		exit 1; \
	fi

.PHONY: all test fmt vet install install-std release-flist archive release-dir get_hub_jwt check-%