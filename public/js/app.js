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
  $routeProvider.when('/login', {templateUrl: 'view/login.html', controller: 'LoginController'});

  $routeProvider.when('/view2', {templateUrl: 'view/partial2.html', controller: 'MyCtrl2'});
  $routeProvider.otherwise({redirectTo: '/'});
}])

.run(function run () {
})

.controller('AppController', ['$scope', '$http', 'Auth', function($scope, $http, Auth) {
	$scope.auth = {
		show: false,
		src: ""
	}
	$scope.user = null;

	$scope.setUser = function (user) {
		$scope.user = user;
	}

	// This function will be called by authiframe when it returns with credentials
	$scope.authorize = function (credentials) {
  		Auth.login(credentials, $scope.setUser);
		$scope.auth.show = false;
		$scope.auth.src = 	"/";
  		$scope.$apply();
	}
  	// Adding this function to window object just to be called by the child iframe
  	window.authorize = function (credentials) {
  		$scope.authorize(credentials);
  	}

	$scope.login = function () {
		$scope.auth.show = true;
		$scope.auth.src = 	"/login";
	}

	$scope.logout = function () {
		Auth.logout();
		$scope.user = null;
	}

	// If user is already logged, lets try to authenticate this user
	Auth.logged($scope.setUser);



	$scope.msg = "No msg at now";

	
}])


;
