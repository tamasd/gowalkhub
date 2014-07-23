(function () {
	"use strict";

	var services = angular.module("gowalkhubServices", ["ngResource"]);

	services.factory("User", ["$resource", function ($resource) {
		return $resource("/api/user/:Id", {"Id": "@Id"}, {
			me: {
				method: "GET",
				params: {
					"Id": "me",
				},
				isArray: false
			},
			save: {
				method: "PUT"
			}
		});
	}]);

	services.factory("Walkthrough", ["$resource", function ($resource) {
		return $resource("/api/v2/walkhub-walkthrough/:uuid", {"uuid": "@uuid"}, {
			save: {
				method: "PUT"
			}
		});
	}]);

	services.factory("Step", ["$resource", function ($resource) {
		return $resource("/api/v2/walkhub-step/:uuid", {"uuid": "@uuid"}, {
			save: {
				method: "PUT"
			}
		});
	}]);
})();
