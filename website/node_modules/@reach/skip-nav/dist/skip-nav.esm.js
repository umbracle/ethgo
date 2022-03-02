import React, { useEffect } from 'react';
import { checkStyles } from '@reach/utils';

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
  useEffect(function () {
    return checkStyles("skip-nav");
  }, []);
  return React.createElement("a", Object.assign({}, props, {
    href: "#" + id,
    "data-reach-skip-link": "",
    "data-reach-skip-nav-link": ""
  }), children);
};

if (process.env.NODE_ENV !== "production") {
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
  return React.createElement("div", Object.assign({}, props, {
    id: id,
    "data-reach-skip-nav-content": ""
  }));
};

if (process.env.NODE_ENV !== "production") {
  SkipNavContent.displayName = "SkipNavContent";
}

export { SkipNavContent, SkipNavLink };
//# sourceMappingURL=skip-nav.esm.js.map
