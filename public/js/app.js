'use strict';


// Declare app level module which depends on filters, and services
angular.module('app', [
  'ngRoute',
  'ngCookies',
  'ngResource',
  'app.filters',
  'app.services',
  'app.directives',
  'app.controllers'
])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/', {templateUrl: 'view/center.html', controller: 'CenterController'});
  $routeProvider.when('/canal/:channelslug', {templateUrl: 'view/channel.html', controller: 'LoginController'});

  $routeProvider.when('/view2', {templateUrl: 'view/partial2.html', controller: 'MyCtrl2'});
  $routeProvider.otherwise({redirectTo: '/'});
}])

.run(function run () {
})

.controller('AppController', ['$scope', '$http', '$timeout', 'Auth', function($scope, $http, $timeout, Auth) {
	$scope.spinner = false;
	$scope.user = null;
	$scope.error = null;
	$scope.channels = null;
	$scope.message = null;

	$scope.newError = function (message) {
		error = { // Creationg a new error object to be returned
			error: {
    			code: 1, // Default error code
    			message: message
    		}
		}
		$scope.showError(erro)
	}

	$scope.showError = function (error) {
		$scope.error = error;
		$timeout($scope.clearError, 5000);
		return
	}

	$scope.clearError = function () {
		$scope.error = null;
	}

	$scope.loginCallback = function (user, error) {
		if (error) {
			$scope.showError(error);
		}
		if (user) {
			$scope.user = user;
		}
	}

	$scope.logout = function () {
		Auth.logout();
		$scope.user = null;
	}

	// If user is already logged, lets try to authenticate this user
	Auth.login($scope.loginCallback);

	// Lets get all channels
	$http({method: 'GET', url: '/channel'}).
	    success(function(data, status, headers, config) {
	    	if (data.error) {
	    		$scope.showError(data)
	    		return
	    	}
        	$scope.channels = data;
	    }).
	    error(function(data, status, headers, config) {
	    	if (data.error) {
	    		$scope.showError(data)
	    		return
	    	}
	    	$scope.newError("Ocorreu um erro ao se buscar a lista de canais.")
    });

}])


;
