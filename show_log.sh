sudo cat /var/log/nginx/access.log | alp --aggregates "users/\d+,products/\d+, products/buy/\d+, comments/\d+, images/" --ave --sum
