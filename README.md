# Century documentation
## Requires
Cassandra database on localhost with created keyspace "century" (I think that automated creation keyspaces is bad practice)

Tested on golang 1.5rc1. Probably it works on versions 1.3-1.5. It doesn`t work on version <1.3.

## Allowed methods

### Create user
```
curl http://localhost:9090/user -i -X POST --data 'login=tester4&password=dfs324'
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Sat, 15 Aug 2015 16:40:59 GMT
Content-Length: 13

{"error":""}
```

### Login
Returns cookie for usage in next requests.
```
curl http://localhost:9090/login -i -X POST --data 'login=tester4&password=dfs324'
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Set-Cookie: sessionToken=dbe23269-bfc9-4ecd-b498-2e50cfed9291
Date: Sat, 15 Aug 2015 16:42:10 GMT
Content-Length: 13

{"error":""}
```
### Check login
```
curl http://localhost:9090/login/check -i -X GET -H 'Cookie: sessionToken=dbe23269-bfc9-4ecd-b498-2e50cfed9291'
HTTP/1.1 200 OK
Date: Sat, 15 Aug 2015 16:43:41 GMT
Content-Length: 27
Content-Type: text/plain; charset=utf-8

{"status":"You logged in!"}
```
### Logout
```
mr_tron@x120e: ~$ curl http://localhost:9090/logout -i -X POST -H 'Cookie: sessionToken=dbe23269-bfc9-4ecd-b498-2e50cfed9291'
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Sat, 15 Aug 2015 16:51:00 GMT
Content-Length: 13

{"error":""}
```
And check login after that.
```
curl http://localhost:9090/login/check -i -X GET -H 'Cookie: sessionToken=dbe23269-bfc9-4ecd-b498-2e50cfed9291'
HTTP/1.1 403 Forbidden
Content-Type: application/json; charset=utf-8
Date: Sat, 15 Aug 2015 16:51:03 GMT
Content-Length: 25

{"error":"unauthorized"}
```

