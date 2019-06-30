#!/bin/sh
curl -H 'Content-Type: application/json' -XDELETE 'localhost:19200/oldindex'
curl -H 'Content-Type: application/json' -XDELETE 'localhost:19200/newindex'

curl -X PUT "localhost:19200/oldindex"
curl -X PUT "localhost:19200/newindex"

curl -X PUT "localhost:19200/oldindex" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "_doc": {
      "properties": {
        "id":   { "type": "text"  },
        "name": { "type": "text"  }
      }
    }
  }
}
'

curl -X PUT "localhost:19200/newindex" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "_doc": {
      "properties": {
        "id":   { "type": "text"  },
        "name": { "type": "text"  }
      }
    }
  }
}
'

# insert
curl -H 'Content-Type: application/json' -X PUT 'localhost:39200/oldindex/_doc/1' -d '{"id":"1","name":"Same Document"}'

curl -H 'Content-Type: application/json' -X PUT 'localhost:39200/newindex/_doc/1' -d '{"id":"1","name":"Same Document"}'
curl -H 'Content-Type: application/json' -X PUT 'localhost:39200/newindex/_doc/2' -d '{"id":"2","name":"New Document"}'
curl -H 'Content-Type: application/json' -X PUT 'localhost:39200/newindex/_doc/3' -d '{"id":"3","name":"New Document 2"}'
