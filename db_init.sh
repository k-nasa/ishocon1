#!/bin/bash
mysql -u root ishocon1 < init/init.sql
ruby init/insert.rb
