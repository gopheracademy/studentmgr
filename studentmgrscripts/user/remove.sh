#!/bin/bash

set -e 

sudo tar jcvf /var/local/userback/$1-home-`date +"%s"`.tar.bz2 /home/$1
sudo userdel -r $1
