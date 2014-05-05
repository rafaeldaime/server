'use strict';

/* Directives */

// It is to add new types of attributes to our html elements
// In this case if we attach the attrybute appVersion, it will alter the text of its element
//
// <span app-version></span>



// I don't think I will use it


angular.module('myApp.directives', []).
  directive('appVersion', ['version', function(version) {
    return function(scope, elm, attrs) {
      elm.text(version);
    };
  }]);