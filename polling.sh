#!/bin/bash

curl -X POST localhost:8050/api/polling >> /var/www/sales/batch.log 2>&1