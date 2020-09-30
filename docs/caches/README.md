# Caches

The cache external resource is an optional addition.
By default, the service don't try to be smart, and will avoid to use any cache.

You choose to have a trade off for your storage system to use a traditional database
that can provide your fact data with cost effectiveness, stability and maintainability perspectives.
But then you don't want to sacrifice the service response times, so you can use a cache system to solve this.
The Caching system do automatic cache invalidation with TTL and on Update/Delete storage operations.

Currently redis and a in-memory implementation is available.

To setup the application to use the cache, either provide the `-cache-url` cli option
or the `CACHE_URL` environment variable.

To setup the cache TTL, you can use the `-cache-ttl` cli option or the `CACHE_TTL` environment variable.
A cache ttl duration in string format must be a unsigned sequence of
decimal numbers, each with optional fraction and a unit suffix,
such as "300ms", "1.5h" or "2h45m".
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

## in-memory

To use the in-memory cache, you have yo set `CACHE_URL` to `memory`.

```bash
export CACHE_URL="memory"
```

the in-memory cache will keep storage results in the memory,
and periodically sync (every ~1m) the cached data with the storage.

Based on the frequency of the access of a certain value, 
the data kept in the memory.

This solution will remove spikes from your storage,
and introduce a more distributed usage on it.

The default `TTL` is `5m`. 
You can modify this trough the `CACHE_TTL`.

The trade-off with this is that your data freshness will suffer from a `1m` latency,
So kill switches and flag release changes will need time before they take effect.

Also keep in mind that this solution use the memory of the web servers,
and they are not synchronized between the web servers.

## redis

This solution allow you to use redis as caching backend.
   
