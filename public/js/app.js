'use strict';


// Declare app level module which depends on filters, and services
var app = angular.module('app', [
  'ngRoute',
  'ngCookies',
  'restangular',
  'angularFileUpload',
  'app.filters',
  'app.services',
  'app.directives',
  'app.controllers'
]);

app.config(function($routeProvider) {
	$routeProvider.when('/', {templateUrl: 'view/center.html', controller: 'CenterController'});
	$routeProvider.when('/canal/:channelslug', {templateUrl: 'view/channel.html', controller: 'CenterController'});

	$routeProvider.otherwise({redirectTo: '/'});
});

app.run(function ($rootScope, $http, $cookies, Restangular) {
	// Cofiguring our Rest client Restangular

	// Calls to api will be done in /api path
	Restangular.setBaseUrl('/api');

	// Configuring id to respect our pattern
	Restangular.configuration.getIdFromElem = function(elem) {
	  // if route is user ==> returns userid
	  return elem[_.initial(elem.route).join('') + "id"];
	}

	// Setting the Api error response handler
	Restangular.setErrorInterceptor(function (response) {
		if (response.status == 401) {
			// The current user is not logged or credentials expired
			console.log("Login required... ");
			if ($cookies.credentials) {
				// If he had credentials, let's remove them
				document.execCommand("ClearAuthenticationCache");
				delete $cookies.credentials;
				Restangular.setDefaultHeaders({Authorization: ""});
				$http.defaults.headers.common.Authorization = "";
				// EXIBIR UMA MENSAGEM AQUI, "VOCE PRECISA SE LOGAR NOVAMENTE"
			} else {
				// EXIBIR UMA MENSAGEM AQUI, "VOCE PRECISA ESTAR LOGADO"
			}
			console.log(response);
		} else if (response.status == 404) {
			console.log("Resource not available...");
			console.log(response);
		} else {
			console.log("Response received with HTTP error code: " + response.status );
			console.log(response);
		}
		return false; // stop the promise chain
	});

	// Adding the function to logout the current user
	$rootScope.logout = function () {
	    document.execCommand("ClearAuthenticationCache");
	    delete $cookies.credentials;
	    Restangular.setDefaultHeaders({Authorization: ""});
	    $http.defaults.headers.common.Authorization = "";
	    delete $rootScope.me;
    }

	// Setting the current user session
	if ($cookies.credentials) {
		// Set the Authorization header saved in this session
		$http.defaults.headers.common.Authorization = $cookies.credentials;
	    Restangular.setDefaultHeaders({Authorization: $cookies.credentials});
	    // Getting the actual user
	    Restangular.one("me").get().then(function (user) {
	    	$rootScope.me = user; // Prevent to create empty object using .$object
	    });
	}
});

app.controller('AppController', function($scope, $timeout, $cookies, Restangular) {
	$scope.error = null;
	$scope.message = null;
	$scope.spinner = false;
	$scope.timeFormat = "'Publicado' d-MM-yy 'às' HH:mm"

	
	$scope.channels = Restangular.all('channels').getList().$object;






	$scope.newMessage = function (message) {
		$scope.message = message;
		$timeout(function () {
			$scope.message = null;
		}, 5000);
	}

	$scope.newError = function (error) {
		if (!error) { // Creating a default error
			$scope.error =  {
				error: {
	    			code: 1,
	    			message: "Ocorreu um erro inesperado!"
	    		}
			}
		} else if (error.error) { // It's an error object
			$scope.error = error;
		}
		else { // New error with the error sent
			$scope.error =  {
				error: {
	    			code: 1,
	    			message: error
	    		}
			}
		}
		$timeout(function () {
			$scope.erro = null;
		}, 5000);
	}

	// Check if there is an error or message in this session
	if ($cookies.message) {
		$scope.newMessage($cookies.message);
		delete $cookies.message;
	}
	if ($cookies.error) {
		$scope.newError($cookies.error);
		delete $cookies.error;
	}



});
