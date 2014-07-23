# GoWalkHub

This is an experimental WalkHub clone with [Revel](http://revel.github.io) and [AngularJS](http://angularjs.org). This is not intended to use in production, if you are interested in WalkHub, see the original product: [WalkHub](https://github.com/Pronovix/WalkHub).

## Installation

### Dependencies

- [go](http://golang.org)
- [npm](https://www.npmjs.org)
- [bower](http://bower.io)
- [gulp](http://gulpjs.com)
- [compass](http://compass-style.org)

### Instructions

After cloning the repository, the first thing you should do is to copy `conf/app.conf.sample` to `conf/app.conf`. Open it in your favourite editor, fill `app.secret` with a randomly generated string. Customize the variables below the `APP settings` comment (db access, Google OAuth2 tokens). 

Then enter the following commands:

```
npm install
bower install
gulp build
revel run github.com/tamasd/gowalkhub
```

You can access the site at http://localhost:3000 or the address what you configured in `conf/app.conf`.

## Development

If you intend to develop, you should make a debug version of the assets by running `gulp --debug`. This compiles everything in debug mode (no minifying) and keeps `gulp` running watching for changes.
