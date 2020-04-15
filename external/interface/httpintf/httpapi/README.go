// Package httpapi provides API on HTTP layer to the toggler service.
//
// The purpose of this application is to provide API over HTTP to toggler service,
// in which you can interact with the service in a programmatic way.
//
//
// Terms Of Service:
//
// there are no TOS at this moment, use at your own risk we take no responsibility
//
//
//    ---
//    BasePath: /api
//    Version: 0.2.0
//
//    Consumes:
//    - application/json
//    Produces:
//    - application/json
//
//    securityDefinitions:
//      AppKey:
//        type: apiKey
//        in: header
//        name: X-APP-KEY
//      AppToken:
//        type: apiKey
//        in: header
//        name: X-APP-TOKEN
//    Security:
//      AppKey: []
//
// swagger:meta
package httpapi
