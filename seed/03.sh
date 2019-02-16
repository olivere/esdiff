#!/bin/sh
curl -H 'Content-Type: application/json' -XDELETE 'localhost:19201/oldindex'
curl -H 'Content-Type: application/json' -XDELETE 'localhost:19201/newindex'
curl -H 'Content-Type: application/json' -XPUT 'localhost:19201/oldindex/event/3' -d '{"id":"3","name":"Same Document"}'

curl -H 'Content-Type: application/json' -XPUT 'localhost:19201/newindex/event/3' -d '{"id":"3","name":"Same Document"}'
curl -H 'Content-Type: application/json' -XPUT 'localhost:19201/newindex/event/2' -d '{"id":"2","name":"New Document"}'
curl -H 'Content-Type: application/json' -XPUT 'localhost:19201/newindex/event/1' -d '{"id":"1","name":"New Document 2"}'
