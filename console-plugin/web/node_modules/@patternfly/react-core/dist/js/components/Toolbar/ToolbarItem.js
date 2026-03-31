"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ToolbarItem = exports.ToolbarItemVariant = void 0;
const tslib_1 = require("tslib");
const React = tslib_1.__importStar(require("react"));
const toolbar_1 = tslib_1.__importDefault(require("@patternfly/react-styles/css/components/Toolbar/toolbar"));
const react_styles_1 = require("@patternfly/react-styles");
const c_toolbar__item_Width_1 = tslib_1.__importDefault(require('@patternfly/react-tokens/dist/js/c_toolbar__item_Width'));
const util_1 = require("../../helpers/util");
const Divider_1 = require("../Divider");
const PageContext_1 = require("../Page/PageContext");
var ToolbarItemVariant;
(function (ToolbarItemVariant) {
    ToolbarItemVariant["separator"] = "separator";
    ToolbarItemVariant["bulk-select"] = "bulk-select";
    ToolbarItemVariant["overflow-menu"] = "overflow-menu";
    ToolbarItemVariant["pagination"] = "pagination";
    ToolbarItemVariant["search-filter"] = "search-filter";
    ToolbarItemVariant["label"] = "label";
    ToolbarItemVariant["chip-group"] = "chip-group";
    ToolbarItemVariant["expand-all"] = "expand-all";
})(ToolbarItemVariant || (exports.ToolbarItemVariant = ToolbarItemVariant = {}));
const ToolbarItem = (_a) => {
    var { className, variant, visibility, spacer, widths, align, alignSelf, alignItems, id, children, isAllExpanded, isOverflowContainer } = _a, props = tslib_1.__rest(_a, ["className", "variant", "visibility", "spacer", "widths", "align", "alignSelf", "alignItems", "id", "children", "isAllExpanded", "isOverflowContainer"]);
    if (variant === ToolbarItemVariant.separator) {
        return React.createElement(Divider_1.Divider, Object.assign({ className: (0, react_styles_1.css)(toolbar_1.default.modifiers.vertical, className) }, props));
    }
    const widthStyles = {};
    if (widths) {
        Object.entries(widths || {}).map(([breakpoint, value]) => (widthStyles[`${c_toolbar__item_Width_1.default.name}${breakpoint !== 'default' ? `-on-${breakpoint}` : ''}`] = value));
    }
    return (React.createElement(PageContext_1.PageContext.Consumer, null, ({ width, getBreakpoint }) => (React.createElement("div", Object.assign({ className: (0, react_styles_1.css)(toolbar_1.default.toolbarItem, variant &&
            toolbar_1.default.modifiers[(0, util_1.toCamel)(variant)], isAllExpanded && toolbar_1.default.modifiers.expanded, isOverflowContainer && toolbar_1.default.modifiers.overflowContainer, (0, util_1.formatBreakpointMods)(visibility, toolbar_1.default, '', getBreakpoint(width)), (0, util_1.formatBreakpointMods)(align, toolbar_1.default, '', getBreakpoint(width)), (0, util_1.formatBreakpointMods)(spacer, toolbar_1.default, '', getBreakpoint(width)), alignItems === 'start' && toolbar_1.default.modifiers.alignItemsStart, alignItems === 'center' && toolbar_1.default.modifiers.alignItemsCenter, alignItems === 'baseline' && toolbar_1.default.modifiers.alignItemsBaseline, alignSelf === 'start' && toolbar_1.default.modifiers.alignSelfStart, alignSelf === 'center' && toolbar_1.default.modifiers.alignSelfCenter, alignSelf === 'baseline' && toolbar_1.default.modifiers.alignSelfBaseline, className) }, (variant === 'label' && { 'aria-hidden': true }), { id: id }, props, (widths && { style: Object.assign(Object.assign({}, widthStyles), props.style) })), children))));
};
exports.ToolbarItem = ToolbarItem;
exports.ToolbarItem.displayName = 'ToolbarItem';
//# sourceMappingURL=ToolbarItem.js.map