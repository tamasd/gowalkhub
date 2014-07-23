(function () {
	"use strict";

	var gulp = require("gulp");

	var args = require("yargs").argv;
	var clean = require("gulp-clean");
	var uglify = require("gulp-uglify");
	var concat = require("gulp-concat");
	var gulpif = require("gulp-if");
	var compass = require("gulp-compass");
	var plumber = require("gulp-plumber");

	var paths = {
		scripts: [
			"bower_components/modernizr/modernizr.js",
			"bower_components/jquery/dist/jquery.js",
			"bower_components/jquery-ui/jquery-ui.js",
			"bower_components/foundation/js/foundation.js",
			"bower_components/angular/angular.js",
			"bower_components/angular-route/angular-route.js",
			"bower_components/angular-animate/angular-animate.js",
			"bower_components/angular-resource/angular-resource.js",
			"bower_components/angular-sanitize/angular-sanitize.js",
			"bower_components/showdown/src/showdown.js",
			"bower_components/angular-markdown-directive/markdown.js",
			"bower_components/showdown/src/extensions/github.js",
			"bower_components/showdown/src/extensions/prettify.js",
			"bower_components/showdown/src/extensions/table.js",
			"bower_components/showdown/src/extensions/twitter.js",

			"js/*.js"
		],
		sass: [
			"sass/*.sass"
		],
		jquery_ui: [
			"bower_components/jquery-ui/themes/cupertino/**"
		]
	};

	gulp.task("buildjs", function () {
		return gulp.src(paths.scripts)
			.pipe(plumber())
			.pipe(concat("app.js"))
			.pipe(gulpif(!args.debug, uglify()))
			.pipe(gulp.dest("public/js"));
	});

	gulp.task("buildsass", function () {
		var sassConfig = {
			css: "public/css",
			sass: "./sass",
			style: "compressed",
			project: __dirname,
			import_path: "bower_components/foundation/scss"
		};

		if (args.debug) {
			sassConfig.style = "expanded";
		}

		return gulp.src(paths.sass)
			.pipe(plumber())
			.pipe(compass(sassConfig));
	});

	gulp.task("copyjqueryui", function () {
		return gulp.src(paths.jquery_ui, { "base": "bower_components/jquery-ui/themes/cupertino" })
			.pipe(gulp.dest("public/jquery-ui"));
	});

	gulp.task("build", ["buildjs", "buildsass", "copyjqueryui"]);

	gulp.task("cleanjs", function () {
		return gulp.src("public/js/app.js")
			.pipe(clean());
	});

	gulp.task("cleansass", function () {
		return gulp.src("public/css/app.css")
			.pipe(clean());
	});

	gulp.task("cleanjqueryui", function () {
		return gulp.src("public/jquery-ui")
			.pipe(clean());
	});

	gulp.task("clean", ["cleanjs", "cleansass", "cleanjqueryui"]);


	gulp.task("watch", function () {
		gulp.watch("js", ["buildjs"]);
		gulp.watch("sass", ["buildsass"]);
	});

	gulp.task("default", ["build", "watch"]);
})();
