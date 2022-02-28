#!/bin/bash
cat /var/log/nginx/access.log | sed 's/|/ /' | awk '{ if ("\"https://hackoween.dev/\"" == $11 || "\"https://www.hackoween.dev/\"" == $11 )print "true" }' | wc -l
