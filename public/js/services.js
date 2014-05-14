'use strict';

/* Services */

// Here we add our:
// service, with injects an instance of the function returned (return new Function() )
// factory, witch return the value that is returned by invoking the function ( return factory() )
// provider, retur the value returned by the $get function defined ( return provider.$get() )


// Demonstrate how to register services
// In this case it is a simple value service.
var services = angular.module('app.services', ['ngResource']);




services.factory('Channels', function ($resource) {
    return $resource('/channel', {}, {
        get: { method: 'GET', isArray: true }
    })
});




// Here we will pass the FullUrl and the ChannelId to a new content be created
services.factory('Content', function ($resource) {
    return $resource('/content/:contentid', {contentid: '@contentid'}, {
        create: {method: 'POST'},
        update: {method: 'PUT'}
    })
});









services.factory('Auth', ['$cookies', '$http', '$resource', function ($cookies, $http, $resource) {
    return {
    	// Get the user in this session and return (user, error)
        session: function (callback) {
        	var that = this;

        	if (!$cookies.credentials) { // Nothing in this session
        		callback(null, null);
        		return
        	}

        	// Set the Authorization header saved in this session
        	$http.defaults.headers.common.Authorization = $cookies.credentials;

        	$resource('/me').get()
        		.$promise.then(function (user) {
        			callback(user, null);
        		}, function (httpResponse) {
		            // Return the error object in response to the caller
		    		callback(null, httpResponse.data);
			    	that.logout();
        		})
        },
        logout: function () {
            document.execCommand("ClearAuthenticationCache");
            delete $cookies.credentials;
            $http.defaults.headers.common.Authorization = "";
        }
    };
}]);
