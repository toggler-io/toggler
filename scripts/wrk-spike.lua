-- example HTTP POST script which demonstrates setting the
-- HTTP method, body, and adding a header

wrk.method = "POST"
wrk.body   = '{ "features": [ "my-feature-flag" ], "id": "public-uniq-pilot-id"}'
wrk.headers["accept"] = "application/json"
wrk.headers["Content-Type"] = "application/json"
