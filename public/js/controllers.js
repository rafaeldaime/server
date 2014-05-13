'use strict';

/* Controllers */

angular.module('app.controllers', [])


  .controller('CenterController', ['$scope', function($scope) {

  }])


  .controller('ContentController', ['$scope', 'Contents', function($scope, Contents) {
  	$scope.usercontent = null;
	$scope.channelid = '';
	$scope.fullurl = "http://";

	// DEBUG !!!
	$scope.debug = function () {
		$scope.channelid = $scope.channels[0].channelid;
		$scope.fullurl = "http://digg.com/tag/news";
	}

	$scope.checkFullUrl = function () {
		// If user paste an URL in the input form that contains the https,
		// than we need to remove de duplicated https.
		if ($scope.fullurl.match(/^(https?:\/\/)(https?:\/\/)/)) {
			$scope.fullurl = $scope.fullurl.replace(/^(https?:\/\/)/, "");
		}
	}

	$scope.clearNewContent = function () {
		$scope.usercontent = null;
		$scope.channelid = '';
		$scope.fullurl = "http://";
	}


	$scope.createContent = function () {
		// If the user don't included the http:// in the url,
		// than we need to include it for him
		if (!$scope.fullurl.match(/^(https?:\/\/)/)) {
			$scope.fullurl = "http://" + $scope.fullurl;
		}
		$scope.spinner = true;
		var newcontent = {
			fullurl: $scope.fullurl,
			channelid: $scope.channelid
		}
		Contents.create(newcontent)
			.$promise.then(function(content) {
				$scope.usercontent = content;
				$scope.fullurl = content.fullurl;
				$scope.channelid = content.channelid;
				$scope.spinner = false;
				$scope.json = "Lol" + angular.toJson($scope.usercontent);
			},function (httpResponse) {
				$scope.newError(httpResponse.data);
				$scope.clearNewContent();
				$scope.spinner = false;
			});
	}

  }]);