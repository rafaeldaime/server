'use strict';

/* Filters */

var filters = angular.module('app.filters', []);

filters.filter('userpicsrc', function() {
	return function(user) {
		if (user) {
			return "pic/" + user.picid + ".png"
		};
		return "pic/default.png";
	};
});


filters.filter('strLimit', function() {
	return function(str, limit) {
		if (str.length > limit) {
			return str.substr(0, limit-3) + " ..."
		};
		return str
	};
});





filters.filter('contentimagesrc', function() {
	return function(content, size) {
		// DUNNO WHY BUY THIS FILTER IS CALLED SO MANY TIMES !!!
		// It's called every $digest action, i dunno if it's the better option...

		//console.log("["+content.contentid+"]Img-size: "+size+" from: "+content.imagemaxsize);
		// Content will always have a small image, even if it's the default
		if (size == 'small') {
			return "img/" + content.imageid + "-" + size + ".png";
		};
		// Check if there is large image for this content, or return default large image
		if (size == 'large' && (content.imagemaxsize == 'small' || content.imagemaxsize == 'medium')) {
			return "img/default-" + size + ".png";
		};
		// Check if there is medium image for this content, or return default medium image
		if (size == 'medium' && content.imagemaxsize == 'small') {
			return "img/default-" + size + ".png";
		};
		return "img/" + content.imageid + "-" + size + ".png";
	};
});
