sudo service nginx restart
sudo service mysqld restart

sudo cp /dev/null /var/log/mysql/slow.log
sudo cp /dev/null /var/log/nginx/access.log
~/benchmark
