#!/bin/bash

# Create the user

PASSWORD=`perl -e 'printf("%s\n", crypt($ARGV[0], "password"))' "$2"`
sudo useradd -g 1002 -m -s /bin/bash -p $PASSWORD $1
sudo usermod --expiredate $(date -d "10 days" "+%Y-%m-%d") $1
echo $1 $2

#create wide startup script
sudo mkdir -p /home/$1/bin/
sudo chown $1:students /home/$1/bin
sudo cp $(dirname $(readlink -f $0))/start-wide.sh /home/$1/bin/
sudo cp $(dirname $(readlink -f $0))/enable-wide.sh /home/$1/bin/
sudo cp /usr/local/bin/wide /home/$1/bin/
sudo chown -R $1:students /home/$1/bin/


#clone wide

sudo git clone https://github.com/bketelsen/wide /home/$1/wide 
sudo chown -R $1:students /home/$1/wide


