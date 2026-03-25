"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.DataList = exports.DataListBase = exports.DataListContext = exports.DataListWrapModifier = void 0;
const tslib_1 = require("tslib");
const React = tslib_1.__importStar(require("react"));
const react_styles_1 = require("@patternfly/react-styles");
const data_list_1 = tslib_1.__importDefault(require("@patternfly/react-styles/css/components/DataList/data-list"));
const gridBreakpointClasses = {
    none: data_list_1.default.modifiers.gridNone,
    always: 'pf-m-grid', // Placeholder per https://github.com/patternfly/patternfly-react/issues/4965#issuecomment-704984236
    sm: data_list_1.default.modifiers.gridSm,
    md: data_list_1.default.modifiers.gridMd,
    lg: data_list_1.default.modifiers.gridLg,
    xl: data_list_1.default.modifiers.gridXl,
    '2xl': data_list_1.default.modifiers.grid_2xl
};
var DataListWrapModifier;
(function (DataListWrapModifier) {
    DataListWrapModifier["nowrap"] = "nowrap";
    DataListWrapModifier["truncate"] = "truncate";
    DataListWrapModifier["breakWord"] = "breakWord";
})(DataListWrapModifier || (exports.DataListWrapModifier = DataListWrapModifier = {}));
exports.DataListContext = React.createContext({
    isSelectable: false
});
const DataListBase = (_a) => {
    var { children = null, className = '', 'aria-label': ariaLabel, onSelectDataListItem, selectedDataListItemId = '', isCompact = false, gridBreakpoint = 'md', wrapModifier = null, onSelectableRowChange, innerRef } = _a, props = tslib_1.__rest(_a, ["children", "className", 'aria-label', "onSelectDataListItem", "selectedDataListItemId", "isCompact", "gridBreakpoint", "wrapModifier", "onSelectableRowChange", "innerRef"]);
    const isSelectable = onSelectDataListItem !== undefined;
    const updateSelectedDataListItem = (event, id) => {
        onSelectDataListItem(event, id);
    };
    return (React.createElement(exports.DataListContext.Provider, { value: {
            isSelectable,
            selectedDataListItemId,
            updateSelectedDataListItem,
            onSelectableRowChange
        } },
        React.createElement("ul", Object.assign({ className: (0, react_styles_1.css)(data_list_1.default.dataList, isCompact && data_list_1.default.modifiers.compact, gridBreakpointClasses[gridBreakpoint], wrapModifier && data_list_1.default.modifiers[wrapModifier], className), style: props.style, role: "list", "aria-label": ariaLabel, ref: innerRef }, props), children)));
};
exports.DataListBase = DataListBase;
exports.DataListBase.displayName = 'DataListBase';
exports.DataList = React.forwardRef((props, ref) => (React.createElement(exports.DataListBase, Object.assign({ innerRef: ref }, props))));
exports.DataList.displayName = 'DataList';
//# sourceMappingURL=DataList.js.map