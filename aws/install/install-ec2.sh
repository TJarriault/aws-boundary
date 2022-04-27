# Installs the boundary as a service for systemd on linux
TYPE=$1
NAME=boundary-DEVOXX

apt-get update
apt-get install apache2 -y

echo "DEMO PAGE for $NAME" > /var/www/html/index.html
