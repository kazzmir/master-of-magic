#!/bin/sh

echo "Building server binary for linux/arm64"
GOOS=linux GOARCH=arm64 go build -trimpath -o server.arm64 .
echo "Copying server binary to remote server"
rsync -avP server.arm64 magic-combat-simulator:server

remote_script=$(cat <<'END_SCRIPT'
echo "Rename old server binary"
sudo mv /opt/magic/server /opt/magic/server.old
echo "Copy new server binary to /opt/magic"
sudo cp server /opt/magic
echo "Update chown"
sudo chown nobody:nogroup /opt/magic/server
echo "Restarting magic service"
sudo systemctl restart magic
echo "Done"
END_SCRIPT
)

ssh magic-combat-simulator "$remote_script"
