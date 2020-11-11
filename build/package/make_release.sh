#!/bin/bash
set -e

#
# Copyright (c) 2020 Jeroen Jacobs/Head In Cloud BV.
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License version 3 as published by
# the Free Software Foundation.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
#

APP=cf-ddns-agent
VERSION=""
BIN_DIR=""
RELEASE_DIR=""


show_usage() {
  echo "Release-builder for cf-ddns-agent"
  echo
  echo "usage:"
  echo "  $0 <bin_dir> <release_dir> <version>"
  echo
  echo "Parameters:"
  echo "  <bin_dir>     : directory which contains the generated platform executables."
  echo "  <release_dir> : directory where generated files should be stored."
  echo "  <version>     : git version tag. If omitted, will be calculated via 'git describe --tags --always --dirty'"

}

prepare_config () {
  BIN_DIR=$(realpath $1)
  RELEASE_DIR=$(realpath $2)
  if [ "$3" = "" ]; then
    VERSION=$(git describe --tags --always --dirty)
  else
    VERSION=$3
  fi

}

build_release() {
  # create dir if not exist
  mkdir -p ${RELEASE_DIR}

  #create zip files for executable in $RELEASE_DIR folder
  echo "** Preparing zip files **"
  for f in $BIN_DIR/*
  do
    FILENAME=`basename $f`
    RELEASE=`echo "$FILENAME" | cut -d'.' -f1`
    EXT=`echo "$FILENAME" | cut -d'.' -f2`
    # if both parts are identical, there was no extension (Posix target).
    if [ "$RELEASE" = "$EXT" ]; then
      cp $f ${RELEASE_DIR}/${APP} ; zip -j ${RELEASE_DIR}/${RELEASE}.zip ${RELEASE_DIR}/${APP} ; rm ${RELEASE_DIR}/${APP}
    else
      # we have an extension (Windows target).
      cp $f ${RELEASE_DIR}/${APP}.${EXT} ; zip -j ${RELEASE_DIR}/${RELEASE}.zip ${RELEASE_DIR}/${APP}.${EXT} ; rm ${RELEASE_DIR}/${APP}.${EXT}
    fi
  done

  #create SHA256 digests
  echo "** Calculating sha256 digests **"
  echo
  pushd $RELEASE_DIR
  rm -f ${APP}_${VERSION}.SHA256
  for f in ./*.zip
  do
    shasum -a 256 `basename $f` >> ${APP}_${VERSION}.SHA256
  done
  cat ${APP}_${VERSION}.SHA256
  shasum -c ${APP}_${VERSION}.SHA256
  popd

  echo "** done **"
}

main () {
  if [ "$#" -lt 2 ] || [ "$#" -gt 3 ]; then
    show_usage
  else
    prepare_config "$@"
    echo "BIN_DIR: $BIN_DIR"
    echo "RELEASE_DIR: $RELEASE_DIR"
    echo "VERSION: $VERSION"
    build_release
  fi
}

main "$@"