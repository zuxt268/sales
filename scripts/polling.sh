#!/bin/bash

curl -X POST localhost:8050/api/growth/polling >> /var/www/sales/batch.log 2>&1