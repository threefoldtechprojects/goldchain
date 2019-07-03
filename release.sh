#!/bin/bash
set -e

package="github.com/nbh-digital/goldchain"

version=$(git describe --abbrev=0)
commit="$(git rev-parse --short HEAD)"

if [ "$commit" == "$(git rev-list -n 1 $version | cut -c1-7)" ]
then
	full_version="$version"
else
	full_version="${version}-${commit}"
fi

ARCHIVE=false
if [ "$1" = "archive" ]; then
	ARCHIVE=true
	shift # remove element from arguments
fi

echo "building version ${version}"

for os in darwin windows linux; do
	echo Packaging ${os}...
	# create workspace
	folder="release/goldchain-${version}-${os}-amd64"
	rm -rf "$folder"
	mkdir -p "$folder"
	# compile binaries
	for pkg in cmd/goldchainc cmd/goldchaind; do
		GOOS=${os} go build -a \
			-ldflags="-X ${package}/pkg/config.rawVersion=${full_version} -s -w" \
			-o "${folder}/${pkg}" "./${pkg}"

	done

	if [ "$ARCHIVE" = true ] ; then
		# add other artifacts
		cp -r LICENSE README.md "$folder"
		# go into the release directory
		pushd release &> /dev/null
		# zip
		(
			zip -rq "goldchain-${version}-${os}-amd64.zip" \
				"goldchain-${version}-${os}-amd64"
		)
		# leave the release directory
		popd &> /dev/null
	fi
done

# create tar.gz to upload as flist
# use previously build linux binaries

# go into release directory
pushd release &> /dev/null

# create directory to contain flist
rm -rf flist
mkdir -p flist

# copy binaries
cp -ar goldchain-${version}-linux-amd64/cmd/* flist/

# make flist
tar -C flist -czvf goldchain-latest.tar.gz ./goldchainc ./goldchaind

# leave release dir
popd &> /dev/null
