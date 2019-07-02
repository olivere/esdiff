# Diff for Elasticsearch

**Warning: This is a work-in-progress. Things might break without warning.**

The `esdiff` tool iterates over two indices in Elasticsearch 5.x or 6.x or 7.x
and performs a diff between the documents in those indices.	

It does so by scrolling over the indices. To allow for a stable sort
order, it uses `_id` by default (`_uid` in ES 5.x).

You need Go 1.11 or later to compile. Install with:

```
$ go install github.com/olivere/esdiff
```

## Example usage

First, we need to setup two Elasticsearch clusters for testing,
then seed a few documents.

```
# Create an Elasticsearch 5.x
# http://localhost:19200 and http://localhost:19201
# Create an Elasticsearch 6.x
# http://localhost:29200 and http://localhost:29201
# Create an Elasticsearch 7.x 
# http://localhost:39200 and http://localhost:39201

$ mkdir -p data

# Increase your docker memory limit (6.0GiB) in Docker App > Preferences > Advanced.
$ docker-compose -f docker-compose.yml up -d

Creating esdiff_elasticsearch5_1 ... done
Creating esdiff_elasticsearch7_1 ... done
Creating esdiff_elasticsearch6_1 ... done

# Check docker containers
$ docker-compose ps
         Name                        Command               State                 Ports
----------------------------------------------------------------------------------------------------
esdiff_elasticsearch5_1   /bin/bash bin/es-docker          Up      0.0.0.0:19200->9200/tcp, 9300/tcp
esdiff_elasticsearch6_1   /usr/local/bin/docker-entr ...   Up      0.0.0.0:29200->9200/tcp, 9300/tcp
esdiff_elasticsearch7_1   /usr/local/bin/docker-entr ...   Up      0.0.0.0:39200->9200/tcp, 9300/tcp

# Check docker container logs 
 ~/src/suhyun/esdiff/ [feature/add_es7*] docker-compose logs -f elasticsearch5
Attaching to esdiff_elasticsearch5_1
elasticsearch5_1  | [2019-07-02T14:17:33,351][WARN ][o.e.b.JNANatives         ] Unable to lock JVM Memory: error=12, reason=Cannot allocate memory
elasticsearch5_1  | [2019-07-02T14:17:33,355][WARN ][o.e.b.JNANatives         ] This can result in part of the JVM being swapped out.
elasticsearch5_1  | [2019-07-02T14:17:33,355][WARN ][o.e.b.JNANatives         ] Increase RLIMIT_MEMLOCK, soft limit: 83968000, hard limit: 83968000
elasticsearch5_1  | [2019-07-02T14:17:33,356][WARN ][o.e.b.JNANatives         ] These can be adjusted by modifying /etc/security/limits.conf, for example:
elasticsearch5_1  | 	# allow user 'elasticsearch' mlockall
........

# Add some documents
# es 5
$ ./seed/es5.sh
# es 6
$ ./seed/es6.sh
# es 7
$ ./seed/es7.sh

# Compile
$ go build
```

Let's make a simple diff:

```
$ ./esdiff -u=true 'http://localhost:39200/newindex/_doc' 'http://localhost:39200/oldindex/_doc'
Updated	1	{*diff.Document}.Source["id"]:
	-: "1"
	+: "2"
{*diff.Document}.Source["name"]:
	-: "Same Document"
	+: "New Document 2"

Deleted	2
```


Notice that you can pass additional options to filter for
the kind of modes that you're interested in. E.g. if you also
want to see all unchanged documents but not those that were
deleted, use `-u=true -d=false`:

```
$ ./esdiff -u=true -d=false 'http://localhost:39200/newindex/_doc' 'http://localhost:39200/oldindex/_doc'
Updated	1	{*diff.Document}.Source["id"]:
	-: "1"
	+: "2"
{*diff.Document}.Source["name"]:
	-: "Same Document"
	+: "New Document 2"
```

Use JSON as output format instead. Together with
[`jq`](https://stedolan.github.io/jq/)
and
[`jiq`](https://github.com/fiatjaf/jiq)
this is quite powerful
(among [other jq-related tools](https://github.com/fiatjaf/awesome-jq)).

```
$ ./esdiff -o json 'http://localhost:39200/newindex/_doc' 'http://localhost:39200/oldindex/_doc' | jq 'select(.mode | contains("deleted"))'
{
  "mode": "deleted",
  "_id": "2",
  "src": {
    "_id": "2",
    "_source": {
      "id": "2",
      "name": "New Document"
    }
  },
  "dst": null
}
```

You can also pass a query to filter the source and/or the destination,
using the `-sf` and `-df` args respectively:

```
$ ./esdiff -o json -sf='{"term":{"user":"olivere"}}' 'http://localhost:39200/newindex/_doc' 'http://localhost:39200/oldindex/_doc' | jq .
{
  "mode": "created",
  "_id": "1",
  "src": null,
  "dst": {
    "_id": "1",
    "_source": {
      "id": "2",
      "name": "New Document 2"
    }
  }
}
```

Use `-h` to display all options:

```
$ ./esdiff -h
General usage:

	esdiff [flags] <source-url> <destination-url>

General flags:
  -a	Print added docs (default true)
  -c	Print changed docs (default true)
  -d	Print deleted docs (default true)
  -df string
    	Raw query for filtering the destination, e.g. {"term":{"name.keyword":"Oliver"}}
  -dsort string
    	Field to sort the destination, e.g. "id" or "-id" (prepend with - for descending)
  -exclude string
    	Raw source filter for excluding certain fields from the source, e.g. "hash_value,sub.*"
  -include string
    	Raw source filter for including certain fields from the source, e.g. "obj.*"
  -o string
    	Output format, e.g. json
  -sf string
    	Raw query for filtering the source, e.g. {"term":{"user":"olivere"}}
  -size int
    	Batch size (default 100)
  -ssort string
    	Field to sort the source, e.g. "id" or "-id" (prepend with - for descending)
  -u	Print unchanged docs
```

## License

MIT. See [LICENSE](https://github.com/olivere/esdiff/blob/master/LICENSE).
