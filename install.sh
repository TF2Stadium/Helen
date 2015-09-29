#!/bin/sh

go build -v

sudo cp ./etc/Helen.conf /etc/Helen.conf
sudo cp ./etc/Helen.service /usr/lib/systemd/system/
sudo cp ./etc/runHelen /usr/bin/runHelen
sudo cp ./Helen /usr/bin/Helen

sudo systemctl daemon-reload
