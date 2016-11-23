#!/bin/bash

cp /opt/caddy/public/etrade.zip /home/$1/

cd /home/$1/ && unzip etrade.zip
cp -R /home/$1/etrade/src/ /home/$1/src/
chown -R $1 /home/$1/src/
