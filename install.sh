#!/bin/sh
set -e

RELEASES_URL="https://github.com/reposaur/reposaur/releases"
FILE_BASENAME="reposaur"

test -z "$VERSION" && VERSION="$(curl -sfL -o /dev/null -w %{url_effective} "$RELEASES_URL/latest" |
  rev |
  cut -f1 -d'/' |
  rev)"

test -z "$VERSION" && {
  echo "Unable to get Reposaur version." >&2
  exit 1
}

test -z "$TEMP_DIR" && TEMP_DIR="$(mktemp -d)"
test -z "$INSTALLATION_DIR" && INSTALLATION_DIR="${HOME}/.reposaur"

mkdir -p "${INSTALLATION_DIR}/bin"

export TAR_FILE="$TEMP_DIR/${FILE_BASENAME}_$(uname -s)_$(uname -m).tar.gz"

echo "Downloading Reposaur ${VERSION}..."
curl -sfLo "$TAR_FILE" \
  "$RELEASES_URL/download/${VERSION}/${FILE_BASENAME}_${VERSION#v}_$(uname -s)_$(uname -m).tar.gz"

echo ""
echo "Installing Reposaur to ${INSTALLATION_DIR}..."
tar -xf "${TAR_FILE}" -C "${INSTALLATION_DIR}"
mv "${INSTALLATION_DIR}/rsr" "${INSTALLATION_DIR}/bin"

# This binary will be deprecated soon
mv "${INSTALLATION_DIR}/reposaur" "${INSTALLATION_DIR}/bin"

echo ""
echo "Done, try running:"
echo "\t${INSTALLATION_DIR}/bin/rsr --help"
echo ""
echo "It's recommended to add the bin directory to your PATH:"
echo "\tPATH=\"\$PATH:${INSTALLATION_DIR}/bin\""
