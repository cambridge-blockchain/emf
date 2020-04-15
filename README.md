
# Overview

Cambridge Blockchain's Echo Microservice Framework (or EMF for short) was designed and built to extend [labstack/echo](https://github.com/labstack/echo) from a simple web framework to an equally simple microservice framework. To achieve this goal, we have added a variety of new functions on the Context object that is passed to the http handler for each request, as well as new middlewares. These features help maintain state and quickly/easily make requests across microservices built using the same framework.

# Additional Features

### Router:
EMF provides a new router, with convience functions for registering vanilla echo or go http handlers, or for directly registering an emf.Handler which takes in an emf.Context and returns an error. Internally, the echo.Router implementation  still does the majority of the heavy lifting.

### RequestHandler:
Vanilla echo does not provide any specific tooling for performing http requests, as it is only focused on serving requests. Therefore, EMF adds an http request handler that wraps go's http.Client to provide easy, consistent, and logged requests to other EMF microservices or any external service. When paired with some properly configured middlewares + other EMF services, the RequestHandler can also pass a RequestID, authorization header, and more through a chain of http calls.

There is a specific function, Requester, that leverages all of EMF's features to take in an EMF service name (not url, see GetDomain() implementation), a request path, input body, and a pointer to JSONDecode the output into. Requester then JSON encodes the input, passes some headers along from the context, performs the request, and JSON decodes the output. If any part in that process fails, it will return an EMFError, otherwise nil. If the request status code is >400, it will attempt to decode the response into the EMFError instead of the output pointer, and return the resulting EMFError.

### EMFError:
Another sticking point when Cambridge Blockchain originally began using EMF was error handling, both returning errors cleanly to the client and between components. To address this, the ErrorHandler functions on emf.Context throw EMFErrors, and the custom error response handler nicely formats and returns these errors over HTTP to the emf.RequestHandler of another EMF service, or any other client.

EMFErrors are configurable via go templates in a yaml configuration file, the path to which is provided when you initialize the EMF service. This allows the endpoints to provide a map[string]interface{} of keys and values to the error handler, and this map is used to populate the error message. This work was designed to also support localization, but some of the functions expect english (patches welcome).

### Models package:
The `models` package exposes many of the most important types and interfaces from across EMF so you have one central package to import for most use cases. Until you need to get into the weeds of customizing EMF functionality, no other packages should be neccessary to import in a file that implements an emf.Handler. The major interfaces for errors and request handling are also exposed here, so if you only intend to use client-side features you should only need `models`.

The package `emf` and its subpackages contains the inner workings of the emf server and context, as well as emf.New for initializing and exposing a server. Many of these packages are neccesary to configure the server and get it properly set up for your needs, but that can all stay in main.go. See example-main.go for a minimal example.

### /info Endpoint:
All EMF services expose a public Info endpoint that provides information about the service, the version of echo and EMF, and the client user agent. This also allows for an easy test that the server is started and the EMF service is running.

For additionally pull-based monitoring, the /metrics endpoint can be enabled to expose prometheus metrics using promauto.

### Integrations / Middlewares:
In addition to the major client features provided by the context, a variety of integrations for monitoring, authentication, logging, and notifications are included. Some of these features were built for the use of specific EMF services built at Cambridge Blockchain, in which case they should be optional / configurable. Some reverse engineering may be neccesary in order to build a comparable Auth or Notifications API but the source should be clear enough and again, patches welcome.

## TODO
The following list outlines important upcoming changes.
- [DOC] Port internal Cambridge Blockchain documentation & How-To's to markdown in [/docs/](docs)
- [DOC] Add a Getting Started guide to this README
- [DOC] Link to godoc, add more function comments w/ examples
- [IMPROVEMENT] Complete Error message localization support

## License

This repository is licensed under the MIT License, as found in the [LICENSE file](LICENSE).
