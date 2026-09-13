#!/bin/sh
set -eu

curl --fail-with-body --request POST http://localhost:8080/orders/update \
  --header 'Content-Type: application/json' \
  --data '{"order_id":"ord_2048","account_id":"acct_7","from":"checkout.paid","to":"fulfillment.packing","message":"Payment cleared; packing started.","event_id":"evt_91"}'
