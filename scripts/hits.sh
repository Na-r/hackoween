#!/bin/bash
cat /var/log/nginx/access.log | sed 's/|/ /' | awk '{print $1}' | sort | wc -l
