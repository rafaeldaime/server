'use strict';

/* Controllers */

angular.module('app.controllers', [])


  .controller('CenterController', ['$scope', function($scope) {

  }])


  .controller('ContentController', ['$scope', 'Content', function($scope, Content) {
  	$scope.usercontent = null;
	$scope.channelid = '';
	$scope.fullurl = "http://";

	// DEBUG !!!
	$scope.debug = function () {
		$scope.fullurl = "http://digg.com/tag/news";
		for (var channelid in $scope.channels) { // get the first
			$scope.channelid = channelid;
			return
		}
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


	// If some attribute in users content has change, lets save it!
	var watchAndSaveContent = function (newValue, oldValue) {
		if (oldValue != null && newValue != null) {
			console.log("Updating the user content!");
			Content.update($scope.usercontent) 
				.$promise.then(function(content) {
					for (var attr in content) {
						//if ($scope.usercontent[attr] != content[attr]) {
						//	$scope.usercontent[attr] = content[attr];
						//}
						$scope.usercontent = content;
					}
					console.log("UPDATED!");
				},function (httpResponse) {
					console.log("SOMETHING HAVE GOT WRONG!");
					$scope.newError(httpResponse.data);
					newValue = oldValue; // Dunno if it works
				});
		}
	}
	// Listen to changes in unsers content
	$scope.$watch("usercontent.title", watchAndSaveContent);
	$scope.$watch("usercontent.description", watchAndSaveContent);


	$scope.createContent = function () {
		// If the user don't included the http:// in the url,
		// than we need to include it for him
		if (!$scope.fullurl.match(/^(https?:\/\/)/)) {
			$scope.fullurl = "http://" + $scope.fullurl;
		}
		$scope.spinner = true;
		var newContentData = {
			FullUrl: $scope.fullurl,
			ChannelId: $scope.channelid
		}
		Content.create(newContentData)
			.$promise.then(function(content) {
				// card instanceof Content
				$scope.usercontent = content;
				$scope.spinner = false;
			},function (httpResponse) {
				$scope.newError(httpResponse.data);
				$scope.clearNewContent();
				$scope.spinner = false;
			});
	}

  }]);