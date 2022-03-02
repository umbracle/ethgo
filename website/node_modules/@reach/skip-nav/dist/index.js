'use strict';

if (process.env.NODE_ENV === 'production') {
  module.exports = require('./skip-nav.cjs.production.min.js');
} else {
  module.exports = require('./skip-nav.cjs.development.js');
}