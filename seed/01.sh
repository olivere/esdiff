#!/bin/sh
curl -H 'Content-Type: application/json' -XDELETE 'localhost:19200/index01'
curl -H 'Content-Type: application/json' -XDELETE 'localhost:29200/index01'
curl -H 'Content-Type: application/json' -XDELETE 'localhost:39200/index01'

# Create mappings
curl -X PUT "localhost:19200/index01" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "tweet": {
      "properties": {
        "user":    { "type": "keyword" },
        "message": { "type": "keyword" }
      }
    }
  }
}
'

curl -X PUT "localhost:29200/index01" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "_doc": {
      "properties": {
        "user":    { "type": "keyword" },
        "message": { "type": "keyword" }
      }
    }
  }
}
'

curl -X PUT "localhost:39200/index01" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "user":    { "type": "keyword" },
      "message": { "type": "keyword" }
    }
  }
}
'

# Add documents
curl -H 'Content-Type: application/json' -XPUT 'localhost:19200/index01/tweet/1' -d '{"user":"olivere","message":"Welcome to Golang"}'
curl -H 'Content-Type: application/json' -XPUT 'localhost:19200/index01/tweet/2' -d '{"user":"olivere","message":"Running is fun"}'
curl -H 'Content-Type: application/json' -XPUT 'localhost:19200/index01/tweet/3' -d '{"user":"sandrae","message":"Playing the piano is fun as well"}'

curl -H 'Content-Type: application/json' -XPUT 'localhost:29200/index01/_doc/1' -d '{"user":"olivere","message":"Welcome to Golang"}'
curl -H 'Content-Type: application/json' -XPUT 'localhost:29200/index01/_doc/3' -d '{"user":"sandrae","message":"Playing the guitar is fun as well"}'
curl -H 'Content-Type: application/json' -XPUT 'localhost:29200/index01/_doc/4' -d '{"user":"sandrae","message":"Climbed that mountain"}'

curl -H 'Content-Type: application/json' -XPUT 'localhost:39200/index01/_doc/1' -d '{"user":"olivere","message":"Welcome to Golang"}'
curl -H 'Content-Type: application/json' -XPUT 'localhost:39200/index01/_doc/3' -d '{"user":"sandrae","message":"Playing the flute, oh boy"}'
curl -H 'Content-Type: application/json' -XPUT 'localhost:39200/index01/_doc/5' -d '{"user":"sandrae","message":"Ran that marathon"}'
