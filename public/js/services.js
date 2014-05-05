'use strict';

/* Services */

// Here we add our:
// service, with injects an instance of the function returned (return new Function() )
// factory, witch return the value that is returned by invoking the function ( return factory() )
// provider, retur the value returned by the $get function defined ( return provider.$get() )


// Demonstrate how to register services
// In this case it is a simple value service.
angular.module('myApp.services', [])

.value('version', '0.1')

.factory('helloWorldFromFactory', function() {
	return {
		sayHello: function() {
			return "Hello, World!"
		}
	}
})



.factory('Auth', ['$cookieStore', '$http', function ($cookieStore, $http) {
    // initialize to whatever is in the cookie, if anything
    $http.defaults.headers.common['Authorization'] = $cookieStore.get('authdata');

    //alert("$cookieStore.get('authdata')="+$cookieStore.get('authdata'))
 
    return {
        setCredentials: function (auth) {
        	//alert("Lets set the credentials:"+auth)
            $http.defaults.headers.common.Authorization = auth;
            $cookieStore.put('authdata', auth);
        },
        clearCredentials: function () {
            document.execCommand("ClearAuthenticationCache");
            $cookieStore.remove('authdata');
            $http.defaults.headers.common.Authorization = 'Basic ';
        }
    };
}])


;
