'use strict';

/* Services */

// Here we add our:
// service, with injects an instance of the function returned (return new Function() )
// factory, witch return the value that is returned by invoking the function ( return factory() )
// provider, retur the value returned by the $get function defined ( return provider.$get() )


// Demonstrate how to register services
// In this case it is a simple value service.
angular.module('app.services', [])

.value('version', '0.1')

.factory('helloWorldFromFactory', function() {
	return {
		sayHello: function() {
			return "Hello, World!"
		}
	}
})



.factory('Auth', ['$cookies', '$http', function ($cookies, $http) {
    return {
    	// Return to callback user, or null if credentials error
        login: function (callback) {
        	var credentials = $cookies.credentials;
        	if (!credentials) {
        		// User not already logged but there is no error
        		callback(null, null);
        		return
        	}
        	this.setCredentials(credentials);
        	var clearCredentials = this.clearCredentials;
        	$http({method: 'GET', url: '/me'}).
			    success(function(data, status, headers, config) {
			    	// OK! Return user
			    	// For some reason, the JSON returned was decodded
			    	if (data.error) {
			    		callback(null, data);
			    		return
			    	}

		        	callback(data);
			    }).
			    error(function(data, status, headers, config) {
			    	if (data.error) { // Returned an APIError
			    		callback(null, data);
			    		clearCredentials();
			    		return
			    	}

	        		error = { // Creationg a new error object to be returned
	        			error: {
		        			code: 1, // Default error code
		        			message: "Ocorreu um erro ao tentar te identificar."
		        		}
	        		}
	        		callback(null, error)
		    		clearCredentials();
			    	
		    });
        },
        logout: function () {
            this.clearCredentials();
        },
        setCredentials: function (credentials) {
            $http.defaults.headers.common.Authorization = credentials;
            $cookies.credentials = credentials;
        },
        clearCredentials: function () {
            document.execCommand("ClearAuthenticationCache");
            delete $cookies.credentials
            $http.defaults.headers.common.Authorization = "";
        }
    };
}])


;
