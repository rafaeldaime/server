'use strict';

/* Directives */

// It is to add new types of attributes to our html elements
// In this case if we attach the attrybute appVersion, it will alter the text of its element
//
// <span app-version></span>



// I don't think I will use it


var directives = angular.module('app.directives', []);


directives.directive('focusMe', function($timeout, $parse) {
  return {
    link: function(scope, element, attrs) {
      var model = $parse(attrs.focusMe);
      scope.$watch(model, function(value) {
        if(value === true) { 
          $timeout(function() {
            element[0].focus(); 
          });
        }
      });
      // on blur event: set attribute value to 'false'
      element.bind('blur', function() {
         scope.$apply(model.assign(scope, false));
      });
    }
  };
});


directives.directive('inputfile', function(){
    return{
        restrict: 'A', // only activate on element attribute
        require: '?ngModel', // get a hold of NgModelController
        link: function($scope, element, attrs, ngModel){
            if(!ngModel) return; // do nothing if no ng-model

            // view -> model
            element.bind('click', function() {
                console.log("Finally clicked on inputfile");
                console.log(element);
            });

            ngModel.$setViewValue(element);

        }
    }
});

directives.directive('contentimageeditable', function(){
    return{
        restrict: 'A', // only activate on element attribute
        require: '?ngModel', // get a hold of NgModelController
        link: function($scope, element, attrs, ngModel){
            element.bind('click', function(){
                angular.element(ngModel.$viewValue).trigger('click');
                console.log('CLICKED ON INPOT CONTENT IMAGE');
                console.log(ngModel.$viewValue);
            });
        }
    }
});


directives.directive('contenteditable', function() {
    return {
		restrict: 'A', // only activate on element attribute
		require: '?ngModel', // get a hold of NgModelController
        link: function($scope, element, attrs, ngModel) {
        	if(!ngModel) return; // do nothing if no ng-model

            // view -> model
            element.bind('blur', function() {
                $scope.$apply(function() {
          			var value = element.html();
                    // Remove tags and etities html
                    value = value.replace(/<(.*?)>|&(.*?);/g, "");
                    // Valide Title
                    value = value.replace(/[^a-zA-Zà-úÀ-Ú0-9 \-.,:;!?]/g, "");
                    // Remove spaces at begin and end of the string
                    value = value.replace(/^\s+|\s+$/g, "");
                    // Replace 2 or more white spaces to just one space
                    value = value.replace(/ {2,}/g, " ");

                    // Value can't be empty, so we don't update them
                    if (value != "") {
                        ngModel.$setViewValue(value);
                    }
                    ngModel.$render(); // Render again the modifications
                });
            });


            // model -> view
            ngModel.$render = function() {
                var value = ngModel.$viewValue;
                element.html(value || '');
            };

            element.bind('keydown', function(event) {
                //console.log("keydown " + event.which);
                var esc = event.which == 27,
                	enter = event.which == 13,
                    elm = event.target;

                if (enter) {
                    // Fields can't have newlines
                    elm.blur();
                    event.preventDefault();
                }

                if (esc) {
                    // Esc returns to the original value
                    ngModel.$render();
                    elm.blur();
                    event.preventDefault();                        
                }
                    
            });
            
        }
    };
});