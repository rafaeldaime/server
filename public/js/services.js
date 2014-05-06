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



.factory('Auth', ['$cookieStore', '$http', function ($cookieStore, $http) {
    return {
    	// Return to callback user, or null if credentials error
        login: function (credentials, callback) {
        	this.setCredentials(credentials);
        	var clearCredentials = this.clearCredentials;
        	$http({method: 'GET', url: '/me'}).
			    success(function(data, status, headers, config) {
			    	// OK! Return user
			    	// For some reason, the JSON returned was decodded
		        	callback(data);
			    }).
			    error(function(data, status, headers, config) {
			    	// ERROR! Alert and return nil
			    	if (data.error) {
			    		alert(data.error.message);
			    	}
			    	else {
			    		alert("Ocorreu um erro ao tentar te identificar")
			    	}
			    	clearCredentials();
			    	callback(null); // User will be setted to null
		    });
        },
        // This function returns to the callback the user already logged
        logged: function (callback) {
        	if ($cookieStore.get('credentials')){
        		this.login($cookieStore.get('credentials'), callback);
        	}
        	else {
        		// User not already logged
        		this.clearCredentials();
        		callback(null);
        	}
        },
        logout: function () {
            this.clearCredentials();
        },
        setCredentials: function (credentials) {
            $http.defaults.headers.common.Authorization = credentials;
            $cookieStore.put('credentials', credentials);
        },
        clearCredentials: function () {
            document.execCommand("ClearAuthenticationCache");
            $cookieStore.remove('credentials');
            $http.defaults.headers.common.Authorization = '';
        }
    };
}])


;
