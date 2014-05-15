'use strict';

/* Services */

// Here we add our:
// service, with injects an instance of the function returned (return new Function() )
// factory, witch return the value that is returned by invoking the function ( return factory() )
// provider, retur the value returned by the $get function defined ( return provider.$get() )


// Demonstrate how to register services
// In this case it is a simple value service.
var services = angular.module('app.services', []);




