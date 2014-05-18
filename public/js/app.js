'use strict';


// Declare app level module which depends on filters, and services
var app = angular.module('app', [
	'ngCookies',
	'ui.router',
	'restangular',
	'angularFileUpload',
	'ui.bootstrap',
	'app.filters',
	'app.services',
	'app.directives',
	'app.controllers'
]);

app.config(function($stateProvider, $urlRouterProvider) {
	// For any unmatched url, redirect to /state1
  	$urlRouterProvider.otherwise("/");
	$stateProvider.state('home', {
			url: "/",
			templateUrl: "view/home.html",
			controller: "HomeController"
		});
});

app.run(function ($rootScope, $http, $cookies, Restangular) {
	// Alert's system
	$rootScope.alerts = [];
	$rootScope.addAlert = function(msg, type) {
		// Type could be danger, success or warning
		if (!type) {
			type = 'warning'
		}
		$rootScope.alerts.push({type: type, msg: msg});
	};
	$rootScope.closeAlert = function(index) {
		$rootScope.alerts.splice(index, 1);
	};

	// Cofiguring our Rest client Restangular

	// Calls to api will be done in /api path
	Restangular.setBaseUrl('/api');

	// Configuring id to respect our pattern
	Restangular.configuration.getIdFromElem = function(elem) {
		// if route is user ==> returns userid
		//console.log("setIdFromElem route: "+elem.route);
		var id = elem["id"];
		var elemid = elem[_.initial(elem.route).join('') + "id"];
		//console.log(elem);
		if (elem["id"]) {
			//console.log("Id: "+elem["id"]);
			return elem["id"]
		}
		//console.log("Id["+_.initial(elem.route).join('') + "id"+"]: "+elem[_.initial(elem.route).join('') + "id"]);
		return elem[_.initial(elem.route).join('') + "id"]
	}

	// Setting the Api error response handler
	Restangular.setErrorInterceptor(function (response) {
		if (response.status === 419) {
			// The users credentials has expired
			// If he had credentials, let's remove them
			document.execCommand("ClearAuthenticationCache");
			delete $cookies.credentials;
			Restangular.setDefaultHeaders({Authorization: ""});
			$http.defaults.headers.common.Authorization = "";
			// Alerting the user
			$rootScope.addAlert("Voce nao esta mais logado no sistema, logue-se novamente.");
		} else if (response.status === 401) {
			// The current user is not logged
			$rootScope.addAlert("Voce nao esta logado no sistema.", "danger");
		} else if (response.status === 404) {
			$rootScope.addAlert("O recurso requisitado nao existe.", "danger");
		} else {
			$rootScope.addAlert("Ops ocorreu um erro.", "danger");
		}
		// JUST FOR DEBBUNGING, LETS SHOW THE ERROR MESSAGE SENT BY SERVER
		if (response.data.error) {
			$rootScope.addAlert(response.data.error.message);
		};
		// return false; // stop the promise chain
	});

	// Adding the function to logout the current user
	$rootScope.logout = function () {
	    document.execCommand("ClearAuthenticationCache");
	    delete $cookies.credentials;
	    Restangular.setDefaultHeaders({Authorization: ""});
	    $http.defaults.headers.common.Authorization = "";
	    delete $rootScope.me;
		$rootScope.addAlert("Voce se deslogou do sistema.");
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


	Restangular.all('categories').getList({order: "categoryname"}).then(function (categories) {
		$rootScope.categories = categories;
	});

});

app.controller('AppController', function($scope, $timeout, $cookies, $modal, $log, Restangular, $controller ) {
	$scope.timeFormat = "d-MM-yy 'às' HH:mm";
	$scope.isSearchHidden = true;
	$scope.isPostHidden = true;
	$scope.isPosting = false;
	$scope.isCategoriesCollapsed = true;

	// Debugging easy

	$scope.postUrl = 'http://digg.com/tag/human-nature';


	// We are waching for categories
	// so we can create the 4 slices of it to show like the 4 columns to user
	var watchCategories = function (newValue, oldValue) {
		if (newValue) {
			var categories = newValue;
			$scope.categoriesSlice1 = categories.slice(0, categories.length/4);
			$scope.categoriesSlice2 = categories.slice(categories.length/4, categories.length/2);
			$scope.categoriesSlice3 = categories.slice(categories.length/2, categories.length*3/4);
			$scope.categoriesSlice4 = categories.slice(categories.length*3/4, categories.length);
		}
	}
	// Listen to changes in categories
	$scope.$watch("categories", watchCategories);


	$scope.editPost = function (content) {
		var modalInstance = $modal.open({
			templateUrl: 'view/newpost.html',
			controller: 'ContentController',
			size: 'lg',
			resolve: {
				categories: function () {
					return $scope.categories;
				},
				content: function () {
					return content;
				}
			}
		});

		modalInstance.result.then(function () { // When user clicked ok
			$scope.addAlert("Seu conteudo foi salvo!", "success");
		}, function () { // When modal is dismissed (lciked out)
			$scope.addAlert("Seu conteudo foi salvo!", "success");
		});
	}

	$scope.newPost = function () {
		$scope.isPosting = true;
		var postUrl = {FullUrl: $scope.postUrl};
		Restangular.all('contents').post(postUrl).then(function (content) {
			$scope.postUrl = "";
			$scope.isPosting = false;
			$scope.isPostHidden = true;
			$scope.editPost(content);
		}, function() { // Ops, there is an error...
			$scope.isPosting = false;
			$scope.isPostHidden = true;
		});
	}



  	// Hidden collapse navbar on click in a link
	$('.nav a').on('click', function(){
		$(".navbar-toggle").click();
		$log.log("Clicked in navbar, close navbar");
	});


});

