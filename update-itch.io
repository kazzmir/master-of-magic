#!/bin/bash

zipfile="Master of Magic.zip"

if [ ! -f "$zipfile" ]; then
    echo "Missing '$zipfile'"
    exit 1
fi

echo "Copying '$zipfile' to data/data.zip"
cp "$zipfile" data/data.zip
make itch.io
echo "Resetting data"
git checkout data
