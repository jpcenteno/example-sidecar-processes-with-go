#! /bin/sh
set -eu

# This is an example preview script that can handle pdf and jpg files.

case "${1}" in
    *.pdf) zathura "${1}" ;;
    *.jpg) imv "${1}" ;;
esac
