/*Package logger is an HTTP middleware for Go that logs web requests to an io.Writer.

  package main

  import (
      "log"
      "net/http"

      "github.com/unrolled/logger"
  )

  var myHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      w.Write([]byte("hello world"))
  })

  func main() {
      loggerMiddleware := logger.New(logger.Options{
          Prefix: "MySampleWebApp",
          RemoteAddressHeaders: []string{"X-Forwarded-Proto"},
          OutputFlags: log.LstdFlags,
      })

      // loggerWithDefaults := logger.New()

      app := loggerMiddleware.Handler(myHandler)
      http.ListenAndServe("0.0.0.0:3000", app)
  }

A simple GET request to "/info/" will output:

  [MySampleWebApp] 2014/11/21 14:11:21 (12.34.56.78) "GET /info/ HTTP/1.1" 200 11 12.54µs

Here's a breakdown of what the values mean:

  [SuppliedPrefix] Date Time (RemoteIP) "Method RequestURI Protocol" StatusCode Size Time
*/
package logger
