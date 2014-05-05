'use strict';

/* Controllers */

angular.module('myApp.controllers', [])

  .controller('LoginController', ['$scope', '$location', 'Auth', function($scope, $location, Auth) {
  	$scope.logged = false;
	$scope.auth = "Nothing yet...";

	var loginframe = angular.element( document.querySelector( '#loginframe' ) );

	$scope.login = function (auth) {
  		Auth.setCredentials(auth);
		this.auth = auth;
  		this.logged = true;
  		this.$apply();
  		//$location.path("/");
	}

  	// ADding this function to window object just to be called by the iframe
  	window.login = function (auth) {
  		$scope.login(auth); // Something not worked fine in this scope. =(
  	}

  }])


  .controller('MainController', ['$scope', '$http', 'Auth', function($scope, $http, Auth) {

  	$scope.me = "You are nothing"

  	$scope.msg = "No msg at now"

  	$scope.logout = function () {
  		Auth.clearCredentials();
  		$scope.msg = "You logged out!"
  	}

  	$scope.getMe = function () {
  		$scope.msg = "We are getting you!"

	  	$http({method: 'GET', url: '/me'}).
		    success(function(data, status, headers, config) {
		      // this callback will be called asynchronously
		      // when the response is available
		      $scope.msg = data;
		    }).
		    error(function(data, status, headers, config) {
		      // called asynchronously if an error occurs
		      // or server returns response with an error status.
		      $scope.msg = data;
	    });
  	}
  }])


  .controller('MyCtrl2', ['$scope', function($scope) {

  }]);