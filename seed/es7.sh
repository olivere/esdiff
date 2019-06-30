#!/bin/sh
curl -H 'Content-Type: application/json' -XDELETE 'localhost:39200/oldindex'
curl -H 'Content-Type: application/json' -XDELETE 'localhost:39200/newindex'

curl -X PUT "localhost:39200/oldindex"
curl -X PUT "localhost:39200/newindex"

# removed _doc type by 7.x
curl -X PUT "localhost:39200/oldindex" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "id":   { "type": "text"  },
      "name": { "type": "text"  }
    }
  }
}
'

curl -X PUT "localhost:39200/newindex" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "id":   { "type": "text"  },
      "name": { "type": "text"  }
    }
  }
}
'

# insert
curl -H 'Content-Type: application/json' -X PUT 'localhost:39200/oldindex/_doc/1' -d '{"id":"1","name":"Same Document"}'
curl -H 'Content-Type: application/json' -X PUT 'localhost:39200/oldindex/_doc/1' -d '{"id":"2","name":"New Document 2"}'

curl -H 'Content-Type: application/json' -X PUT 'localhost:39200/newindex/_doc/1' -d '{"id":"1","name":"Same Document"}'
curl -H 'Content-Type: application/json' -X PUT 'localhost:39200/newindex/_doc/2' -d '{"id":"2","name":"New Document"}'
curl -H 'Content-Type: application/json' -X PUT 'localhost:39200/newindex/_doc/3' -d '{"id":"3","name":"New Document 2"}'
