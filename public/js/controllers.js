'use strict';

/* Controllers */

controllers = angular.module('app.controllers', []);


controllers.controller('HomeController', function($scope, Restangular) {
	var getRows = function(array, columns) {
		var rows = [];
		var i, j, temparray, chunk = columns;
		for (i=0,j=array.length; i<j; i+=chunk) {
			temparray = array.slice(i, i+chunk);
			rows.push(temparray);
		}
		return rows;
	};

	Restangular.all('contents').getList({order: "top"}).then(function (contents) {
		console.log(contents);
		$scope.contentrows = getRows(contents, 3);
	});

	$scope.like = function (content) {
		console.log("Liking: "+content.contentid)

		content.all("likes").post()
			.then(function (returndata) {
				console.log("Like added");
				console.log("ilikedike count="+returndata.ilike);
				content.likecount = returndata.likecount;
				content.ilike = returndata.ilike;
			});
	}

	$scope.unlike = function (content) {
		console.log("Unliking: "+content.contentid)

		content.all("likes").remove()
			.then(function (returndata) {
				console.log("Like removed");
				console.log("ilikedike count="+returndata.ilike);
				content.likecount = returndata.likecount;
				content.ilike = returndata.ilike;
			});
	}
});






var ContentController = controllers.controller('ContentController', function($scope, Restangular, $upload, $modalInstance, categories, content) {
	$scope.content = content;

	$scope.ok = function () {
		$modalInstance.close();
	};


	$scope.$watch("content", function (newValue, oldValue) {
		// Checks if it isn't a brand new content
		if (oldValue && oldValue.cotentid == newValue.cotentid) {
			// Let's check if user changed title, description or category
			if (oldValue.title != newValue.title ||
				oldValue.description != newValue.description ||
				oldValue.categoryid != newValue.categoryid) {
				// We should save it's new value
				newValue.save();
				console.log("New value saved!");
			}
		}
	}, true); // Object equality (not just reference).


	$scope.onFileSelect = function($files) {
	    //$files: an array of files selected, each file has name, size, and type.
	    for (var i = 0; i < $files.length; i++) {
	      var file = $files[i];
	      $scope.upload = $upload.upload({
	        url: 'api/contents/'+$scope.content.contentid+'/image', //upload.php script, node.js route, or servlet url
	        method: 'POST',
	        // headers: {'header-key': 'header-value'},
	        // withCredentials: true,
	        //data: $scope.content, //{myObj: $scope.myModelObj},
	        file: file, // or list of files: $files for html5 only
	        /* set the file formData name ('Content-Desposition'). Default is 'file' */
	        //fileFormDataName: myFile, //or a list of names for multiple files (html5).
	        /* customize how data is added to formData. See #40#issuecomment-28612000 for sample code */
	        //formDataAppender: function(formData, key, val){}
	      }).progress(function(evt) {
	        console.log('percent: ' + parseInt(100.0 * evt.loaded / evt.total));
	      }).success(function(content, status, headers, config) {
	        // file is uploaded successfully
	        console.log("Sucess!");
	        console.log(content);
	        $scope.content.imageid = content.imageid;
	      }).error(function(data, status, headers, config) {
	        // file is uploaded successfully
	        console.log("ERRO!");
	        console.log(data);
	      });
	      //.then(success, error, progress); 
	      //.xhr(function(xhr){xhr.upload.addEventListener(...)})// access and attach any event listener to XMLHttpRequest.
	    }
	}

  });