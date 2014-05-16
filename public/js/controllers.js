'use strict';

/* Controllers */

angular.module('app.controllers', [])


  .controller('CenterController', function($scope) {

  })


  .controller('ContentController', function($scope, Restangular, $upload) {
  	$scope.usercontent = null;
	$scope.channelid = '';
	$scope.fullurl = "http://";


	// DEBUG !!!
	$scope.debug = function () {
		$scope.fullurl = "http://digg.com/tag/news";
		$scope.channelid = _.first($scope.channels).channelid;
	}

	$scope.checkFullUrl = function () {
		// If user paste an URL in the input form that contains the https,
		// than we need to remove de duplicated https.
		if ($scope.fullurl.match(/^(https?:\/\/)(https?:\/\/)/)) {
			$scope.fullurl = $scope.fullurl.replace(/^(https?:\/\/)/, "");
		}
	}

	$scope.closeContent = function () {
		$scope.usercontent = null;
		$scope.channelid = '';
		$scope.fullurl = "http://";
	}


	// If some attribute in users content has change, lets save it!
	var watchAndSaveContent = function (newValue, oldValue) {
		if (oldValue != null && newValue != null) {
			console.log("Updating the user content!");
			$scope.usercontent.save();
		}
	}
	// Listen to changes in unsers content
	$scope.$watch("usercontent.title", watchAndSaveContent);
	$scope.$watch("usercontent.description", watchAndSaveContent);
	$scope.$watch("usercontent.channelid", watchAndSaveContent);


	$scope.createContent = function () {
		$scope.spinner = true;
		var newContentData = {
			FullUrl: $scope.fullurl,
			ChannelId: $scope.channelid
		}
		Restangular.all('contents').post(newContentData).then(function (content) {
				$scope.usercontent = content;
				$scope.spinner = false;
		});
	}





	$scope.triggerImageUpload = function () {
		var imageInput = angular.element(document.querySelector( '#usercontent-imageinput' ));
	    imageInput.trigger('click');
	}

	$scope.onFileSelect = function($files) {
	    //$files: an array of files selected, each file has name, size, and type.
	    for (var i = 0; i < $files.length; i++) {
	      var file = $files[i];
	      $scope.upload = $upload.upload({
	        url: 'api/contents/'+$scope.usercontent.contentid+'/image', //upload.php script, node.js route, or servlet url
	        method: 'POST',
	        // headers: {'header-key': 'header-value'},
	        // withCredentials: true,
	        //data: $scope.usercontent, //{myObj: $scope.myModelObj},
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
	        $scope.usercontent.imageid = content.imageid;
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