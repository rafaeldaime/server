'use strict';

/* Directives */

// It is to add new types of attributes to our html elements
// In this case if we attach the attrybute appVersion, it will alter the text of its element
//
// <span app-version></span>



// I don't think I will use it


var directives = angular.module('app.directives', []);

directives.directive('contenteditable', function() {
    return {
		restrict: 'A', // only activate on element attribute
		require: '?ngModel', // get a hold of NgModelController
        link: function(scope, element, attrs, ngModel) {
        	if(!ngModel) return; // do nothing if no ng-model

            // view -> model
            element.bind('blur', function() {
                //console.log("BIDDING!");
                scope.$apply(function() {
          			var newValue = element.html();
          			if( attrs.removeBr ) {
			            newValue = newValue.replace(/<br\s*[\/]?>/gi, "");
			            element.html(newValue)
			        }
          			if( attrs.stripBr ) {
			            newValue = newValue.replace(/<br\s*[\/]?>/gi, "\n");
			        }
                    ngModel.$setViewValue(newValue);
                });
            });

            // model -> view
            ngModel.$render = function() {
                //console.log("RENDER!");
                var value = ngModel.$viewValue;
      			if( value && attrs.stripBr ) {
		            value = value.replace(/\\n/gi, "<br>");
		        }
                element.html(value || '');
            };

            element.bind('keydown', function(event) {
                //console.log("keydown " + event.which);
                var esc = event.which == 27,
                	enter = event.which == 13,
                    elm = event.target;

                if (enter) {
                    //console.log("enter");
          			if( attrs.removeBr ) {
	                    elm.blur();
	                    event.preventDefault(); 
			        }                        
                }

                if (esc) {
                    //console.log("esc");
                    elm.blur();
                    event.preventDefault();                        
                }
                    
            });
            
        }
    };
});