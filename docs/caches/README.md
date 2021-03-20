# Caches

If the backing storage you use is not able to handle the load on it's own,
you can enable caching in toggler.

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
 
The default `TTL` is `5m`. 

## in-memory

To use the in-memory cache, you have yo set `CACHE_URL` to `memory`.

```bash
export CACHE_URL="memory"
```

the in-memory cache will keep storage results in the memory,
and subscribe to any event that would invalidate the cached data compared to the storage.

Based on the frequency of the access of a certain value, the data kept in the memory.

This solution will remove spikes from your storage,
and introduce a more distributed usage on it.

Keep in mind that this solution use the memory of the web servers,
and they are not synchronized proactively between the web servers.

## redis

This solution allow you to use redis as caching backend.
