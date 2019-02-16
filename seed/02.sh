#!/bin/sh
curl -H 'Content-Type: application/json' -XDELETE 'localhost:19200/oldindex'
curl -H 'Content-Type: application/json' -XDELETE 'localhost:19200/newindex'
curl -H 'Content-Type: application/json' -XPUT 'localhost:19200/oldindex/event/239473748' -d '{"id":"239473748","name":"Same Document"}'

curl -H 'Content-Type: application/json' -XPUT 'localhost:19200/newindex/event/239473748' -d '{"id":"239473748","name":"Same Document"}'
curl -H 'Content-Type: application/json' -XPUT 'localhost:19200/newindex/event/34' -d '{"id":"34","name":"New Document"}'
curl -H 'Content-Type: application/json' -XPUT 'localhost:19200/newindex/event/32' -d '{"id":"32","name":"New Document 2"}'
