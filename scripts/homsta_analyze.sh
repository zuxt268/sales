#!/bin/bash

curl -X POST localhost:8050/api/analyze/domains >> /var/www/sales/batch.log 2>&1