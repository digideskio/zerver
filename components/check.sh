#!/bin/sh
curl -X GET -H "X-XSRFToken:$1" http://127.0.0.1:4001/access -v
