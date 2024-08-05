#! /bin/sh
set -eu

case "${1}" in
    *.pdf) zathura "${1}" ;;
    *.jpg) imv "${1}" ;;
esac
