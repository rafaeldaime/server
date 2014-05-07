'use strict';

/* Directives */

// It is to add new types of attributes to our html elements
// In this case if we attach the attrybute appVersion, it will alter the text of its element
//
// <span app-version></span>



// I don't think I will use it


angular.module('app.directives', [])

.directive('auth', function () {       
    return {
        link: function(scope, element, attrs) {   

            element.bind("load" , function(e){ 

                // success, "onload" catched
                // now we can do specific stuff:

                //alert('Auth carregou');
                scope.auth.show = true;
                scope.spinner = false;

                if(element[0].naturalHeight > element[0].naturalWidth){
                    element.removeClass("horizontal").addClass("vertical");
                }
            });

        }
    }
})


.directive('appVersion', ['version', function(version) {
	return function(scope, elm, attrs) {
		elm.text(version);
	};
}])


;