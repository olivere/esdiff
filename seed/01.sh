#!/bin/sh
curl -H 'Content-Type: application/json' -XPUT 'localhost:19200/index01/tweet/1' -d '{"user":"olivere","message":"Welcome to Golang"}'
curl -H 'Content-Type: application/json' -XPUT 'localhost:19200/index01/tweet/2' -d '{"user":"olivere","message":"Running is fun"}'
curl -H 'Content-Type: application/json' -XPUT 'localhost:19200/index01/tweet/3' -d '{"user":"sandrae","message":"Playing the piano is fun as well"}'

curl -H 'Content-Type: application/json' -XPUT 'localhost:19201/index01/_doc/1' -d '{"user":"olivere","message":"Welcome to Golang"}'
curl -H 'Content-Type: application/json' -XPUT 'localhost:19201/index01/_doc/3' -d '{"user":"sandrae","message":"Playing the guitar is fun as well"}'
curl -H 'Content-Type: application/json' -XPUT 'localhost:19201/index01/_doc/4' -d '{"user":"sandrae","message":"Climbed that mountain"}'
