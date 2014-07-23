(function () {
	"use strict";

	var controllers = angular.module("gowalkhubControllers", []);

	controllers.controller("MainPageController", ["$scope", "$http", function ($scope, $http) {
		$scope.walkthroughs = [];

		$http.get("/api/walkthrough")
			.success(function (data) {
				$scope.walkthroughs = data;
			});
	}]);

	controllers.controller("ProfileController", ["$scope", "$routeParams", "User", function ($scope, $routeParams, User) {
		$scope.profileEdit = false;
		$scope.profileCanEdit = false;

		$scope.UserRoles = {};
		$scope.UserRoles[$scope.UserRoleDisabled] = "Disabled";
		$scope.UserRoles[$scope.UserRoleNormal] = "Normal";
		$scope.UserRoles[$scope.UserRoleAdmin] = "Admin";

		User.get({Id: $routeParams.Id}, function (data) {
			$scope.user = angular.copy(data);
			$scope.originalUserData = angular.copy(data);

			$scope.profileCanEdit = $scope.currentUser.Id === $scope.user.Id || $scope.currentUser.Role == $scope.UserRoleAdmin;
		});

		$scope.startEdit = function () {
			$scope.profileEdit = true;
		};

		$scope.save = function () {
			$scope.user.$save();
			$scope.profileEdit = false;
		};

		$scope.cancel = function () {
			$scope.user = angular.copy($scope.originalUserData);
			$scope.profileEdit = false;
		};
	}]);

	controllers.controller("WTRecordController", ["$scope", "$http", "$location", function ($scope, $http, $location) {
		$scope.title = "";
		$scope.url = "";
		$scope.use_proxy = true;

		$scope.walkhub_proxy_url = window.WALKHUB_PROXY_URL;

		$scope.startDrupal();

		$scope.save = function () {
			var steps = $("#edit-steps").val();
			var data = {
				Title: $scope.title || "Untitled walkthrough",
				PasswordParameter: $("#edit-password-parameter").is(":checked"),
				Steps: steps ? JSON.parse(steps) : []
			};

			$http.post("/api/walkthrough", data)
				.success(function (data) {
					$location.path("/walkthrough/" + data.uuid);
				})
				.error(function () {
					alert("Failed to save walkthrough.");
				});
		};
	}]);

	controllers.controller("WTPlayController", ["$scope", "$routeParams", "Walkthrough", "Step", function ($scope, $routeParams, Walkthrough, Step) {
		$scope.walkthroughEdit = false;
		$scope.stepEdit = {};

		$scope.walkthrough = {};
		$scope.steps = {};
		$scope.originalWalkthrough = null;
		$scope.originalSteps = {};


		Walkthrough.get({"uuid": $routeParams.uuid}, function (data) {
			$scope.walkthrough = angular.copy(data);
			$scope.originalWalkthrough = angular.copy(data);

			for (var i in data.steps) {
				if (data.steps.hasOwnProperty(i)) {
					$scope.stepEdit[data.steps[i]] = false;
					Step.get({"uuid": data.steps[i]}, function (data) {
						$scope.steps[data.uuid] = angular.copy(data);
						$scope.originalSteps[data.uuid] = angular.copy(data);
					});
				}
			}

			$scope.walkthroughCanEdit = $scope.walkthrough.author === $scope.currentUser.Id || $scope.currentUser.Role === UserRoleAdmin;

			// TODO use this instead: http://lorenzmerdian.blogspot.de/2013/03/how-to-handle-dom-updates-in-angularjs.html
			setTimeout($scope.startDrupal, 500);
		});

		$scope.startEdit = function () {
			$scope.walkthroughEdit = true;
		};

		$scope.save = function () {
			$scope.walkthrough.$save();
			$scope.walkthroughEdit = false;
		};

		$scope.cancel = function () {
			$scope.walkthrough = angular.copy($scope.originalWalkthrough);
			$scope.walkthroughEdit = false;
		};

		$scope.startStepEdit = function (uuid) {
			$scope.stepEdit[uuid] = true;
		};

		$scope.saveStep = function (uuid) {
			$scope.steps[uuid].$save(null, function (data) {
				$scope.steps[uuid] = angular.copy(data);
				$scope.originalSteps[uuid] = angular.copy(data);
			});
			$scope.stepEdit[uuid] = false;
		};

		$scope.cancelStep = function (uuid) {
			$scope.steps[uuid] = angular.copy($scope.originalSteps[uuid]);
			$scope.stepEdit[uuid] = false;
		};
	}]);
})();
