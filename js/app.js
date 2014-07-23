(function () {
	"use strict";

	$(document).foundation();

	var gowalkhubApp = angular.module("gowalkhub", [
		"ngRoute",
		"ngSanitize",
		"btford.markdown",
		"gowalkhubControllers",
		"gowalkhubServices"
	]);

	gowalkhubApp.run(["$rootScope", "User", "$http", function ($rootScope, User, $http) {
		$rootScope.UserRoleDisabled = 0;
		$rootScope.UserRoleNormal = 1;
		$rootScope.UserRoleAdmin = 2;

		$rootScope.startDrupal = function () {
			Drupal.behaviors.walkhub.attach($("#main-container"));
		};

		$rootScope.navbar = true;
		User.me(function (me) {
			$rootScope.currentUser = me;
			$rootScope.loggedIn = !!me.Id;
			$rootScope.userLoaded = true;
		}, function () {
			$rootScope.userLoaded = true;
		});

		$http.defaults.headers.common["X-CSRF-Token"] = window.CSRF_TOKEN;
		$rootScope.csrfToken = window.CSRF_TOKEN;
	}]);

	gowalkhubApp.config(["$routeProvider", "$locationProvider", "markdownConverterProvider", function ($routeProvider, $locationProvider, markdownConverterProvider) {
		$locationProvider.html5Mode(true);
		$locationProvider.hashPrefix("!");

		markdownConverterProvider.config({
			extensions: ["github", "prettify", "table", "twitter"]
		});

		$routeProvider
			.when("/walkthrough/record", {
				templateUrl: "/public/partials/wtrecord.html",
				controller: "WTRecordController"
			})
			.when("/walkthrough/:uuid", {
				templateUrl: "/public/partials/wtplay.html",
				controller: "WTPlayController"
			})
			.when("/user/:Id", {
				templateUrl: "/public/partials/profile.html",
				controller: "ProfileController"
			})
			.when("/", {
				templateUrl: "/public/partials/mainpage.html",
				controller: "MainPageController"
			});
	}]);

	gowalkhubApp.filter("playButton", function () {
		return function (walkthrough) {
			var a = document.createElement("a");	
			a.href = "#";
			a.innerHTML = "Play walkthrough";
			a.setAttribute("class", "walkthrough-start button tiny");

			a.setAttribute("data-walkthrough-proxy-url", window.WALKHUB_PROXY_URL);
			a.setAttribute("data-walkthrough-uuid", walkthrough.uuid);

			for (var p in walkthrough.parameters) {
				if (walkthrough.parameters.hasOwnProperty(p)) {
					a.setAttribute("data-walkthrough-parameter-" + p, walkthrough.parameters[p]);
				}
			}

			// @TODO embedjs support
			a.setAttribute("data-embedjs", "");
			a.setAttribute("data-embedjskey", "");

			a.setAttribute("data-social-sharing", "1");

			a.setAttribute("data-walkthrough-edit-link", window.location.href);

			// @TODO add support for severities
			a.setAttribute("data-walkthrough-severity", "2");
			a.setAttribute("data-walkthrough-severity-text", "This walkthrough: <em class=\"placeholder\">changes configuration (run only on staging or test environments)</em>");

			return a.outerHTML;
		};
	});

	gowalkhubApp.filter("trust", ["$sce", function ($sce) {
		return function (text) {
			return $sce.trustAsHtml(text);
		};
	}]);

	gowalkhubApp.filter("step", function () {
		return function (step) {
			var out = "";

			if (!step) {
				return "";
			}

			out += step.command + "(";

			if (step.arg1) {
				out += step.arg1;

				if (step.arg2) {
					out += ", " + step.arg2;
				}
			}

			out += ")";

			if (step.highlight) {
				out += " :: " + step.highlight;
			}
			
			return out;
		};
	});

})();
