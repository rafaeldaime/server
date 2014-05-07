'use strict';

/* Filters */

angular.module('app.filters', [])

.filter('userpicsrc', function() {
	return function(user) {
		if (user) {
			return "img/pic/" + user.picid + ".png"
		};
		return "img/pic/default.png";
	};
})


.filter('interpolate', ['version', function(version) {
	return function(text) {
		return String(text).replace(/\%VERSION\%/mg, version);
	};
}])


 ;