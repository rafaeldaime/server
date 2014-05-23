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
	
	app.$stateProvider = $stateProvider;
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
	// Verify if there is an message or error stored in cookies
	if ($cookies.message) {
		$rootScope.addAlert($cookies.message)
		delete $cookies.message
	}
	if ($cookies.error) {
		$rootScope.addAlert($cookies.error)
		delete $cookies.error
	}

	// Helper function to divide array in a 2d array
	// Return an array with max of elements in each row
	$rootScope.getRows = function(array, maxElem) {
		var rows = [];
		var i, j, temparray, chunk = maxElem;
		for (i=0,j=array.length; i<j; i+=chunk) {
			temparray = array.slice(i, i+chunk);
			rows.push(temparray);
		}
		return rows;
	};
	// Helper function to divide array in a 2d array
	// Divide the inner array in x pieces
	$rootScope.getCols = function(array, pieces) {
		var rows = [];
		var i, j, temparray, chunk = array.length/pieces;
		for (i=0,j=array.length; i<j; i+=chunk) {
			temparray = array.slice(i, i+chunk);
			rows.push(temparray);
		}
		return rows;
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
		

		// Configuring our router to add each category
		_.forEach(categories, function (category) {
			//console.log("Adding category "+category.categoryslug+" to router.");
			app.$stateProvider.state(category.categoryslug, {
				url: "/"+category.categoryslug,
				templateUrl: "view/home.html",
				controller: "HomeController",
				data: {
					categoryslug: category.categoryslug
				}
			});
		})
		

		// Attaching some helper functions
		categories.getName = function (content) {
			if (content) {
				return _.find(this, function (category) {
					return category.categoryid == content.categoryid
				}).categoryname
			};
		}
		categories.getSlug = function (content) {
			if (content) {
				return _.find(this, function (category) {
					return category.categoryid == content.categoryid
				}).categoryslug
			};
		}

		// Attaching to root scope, cause all controllers will need it
		$rootScope.categories = categories;
		$rootScope.categoriesCols = $rootScope.getCols(categories, 4);
	});
});

app.controller('AppController', function($scope, $timeout, $cookies, $modal, $log, Restangular, $controller ) {
	$scope.timeFormat = "d-MM-yy 'às' HH:mm";
	$scope.showSearch = false;
	$scope.isSearching = false;
	$scope.showPost = false;
	$scope.isPosting = false;
	$scope.postUrl = "";



	$scope.editPost = function (content) {
		var modalInstance = $modal.open({
			templateUrl: 'view/postedit.html',
			controller: 'ContentController',
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

