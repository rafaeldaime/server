'use strict';


// Declare app level module which depends on filters, and services
var app = angular.module('app', [
  'ngRoute',
  'ngCookies',
  'ngResource',
  'app.filters',
  'app.services',
  'app.directives',
  'app.controllers'
]);

app.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/', {templateUrl: 'view/center.html', controller: 'CenterController'});
  $routeProvider.when('/canal/:channelslug', {templateUrl: 'view/channel.html', controller: 'CenterController'});

  $routeProvider.otherwise({redirectTo: '/'});
}]);

app.run(function run () {
});

app.controller('AppController', ['$scope', '$http', '$timeout', '$cookies', 'Auth', 'Channels', function($scope, $http, $timeout, $cookies, Auth, Channels) {
	$scope.error = null;
	$scope.message = null;
	$scope.spinner = false;
	$scope.timeFormat = "d-MM-yy 'às' HH:mm"

	$scope.user = null;
	$scope.channels = {};


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


	// Taking the user present in this session or nothing
	// obs, if there is no user, callback will be called with (null, null)
	Auth.session(function (user, error) {
		if (user) {
			$scope.user = user;
		}
		if (error) {
			$scope.newError(error);
		}
	});

	$scope.logout = function () {
		Auth.logout();
		$scope.user = null;
	}





	Channels.get()
		.$promise.then(function(channels) {
			for (var i in channels) {
				$scope.channels[ channels[i].channelid ] = channels[i];
			}
			//$scope.channels = channels;
		},function (httpResponse) { // If server returned an error...
			$scope.newError(httpResponse.data);
		});



}]);
