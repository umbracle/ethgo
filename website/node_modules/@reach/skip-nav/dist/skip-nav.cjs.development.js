'use strict';

Object.defineProperty(exports, '__esModule', { value: true });

var React = require('react');
var utils = require('@reach/utils');

function _interopDefaultLegacy (e) { return e && typeof e === 'object' && 'default' in e ? e : { 'default': e }; }

var React__default = /*#__PURE__*/_interopDefaultLegacy(React);

function _objectWithoutPropertiesLoose(source, excluded) {
  if (source == null) return {};
  var target = {};
  var sourceKeys = Object.keys(source);
  var key, i;

  for (i = 0; i < sourceKeys.length; i++) {
    key = sourceKeys[i];
    if (excluded.indexOf(key) >= 0) continue;
    target[key] = source[key];
  }

  return target;
}

// menus on a page a use might want to skip at various points in tabbing?).

var defaultId = "reach-skip-nav"; ////////////////////////////////////////////////////////////////////////////////

/**
 * SkipNavLink
 *
 * Renders a link that remains hidden until focused to skip to the main content.
 *
 * @see Docs https://reach.tech/skip-nav#skipnavlink
 */

var SkipNavLink = function SkipNavLink(_ref) {
  var _ref$children = _ref.children,
      children = _ref$children === void 0 ? "Skip to content" : _ref$children,
      contentId = _ref.contentId,
      props = _objectWithoutPropertiesLoose(_ref, ["children", "contentId"]);

  var id = contentId || defaultId;
  React.useEffect(function () {
    return utils.checkStyles("skip-nav");
  }, []);
  return React__default['default'].createElement("a", Object.assign({}, props, {
    href: "#" + id,
    "data-reach-skip-link": "",
    "data-reach-skip-nav-link": ""
  }), children);
};

{
  SkipNavLink.displayName = "SkipNavLink";
} ////////////////////////////////////////////////////////////////////////////////

/**
 * SkipNavContent
 *
 * Renders a div as the target for the link.
 *
 * @see Docs https://reach.tech/skip-nav#skipnavcontent
 */


var SkipNavContent = function SkipNavContent(_ref2) {
  var idProp = _ref2.id,
      props = _objectWithoutPropertiesLoose(_ref2, ["id"]);

  var id = idProp || defaultId;
  return React__default['default'].createElement("div", Object.assign({}, props, {
    id: id,
    "data-reach-skip-nav-content": ""
  }));
};

{
  SkipNavContent.displayName = "SkipNavContent";
}

exports.SkipNavContent = SkipNavContent;
exports.SkipNavLink = SkipNavLink;
//# sourceMappingURL=skip-nav.cjs.development.js.map
