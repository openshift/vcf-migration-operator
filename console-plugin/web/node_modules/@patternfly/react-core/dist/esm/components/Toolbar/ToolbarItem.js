import { __rest } from "tslib";
import * as React from 'react';
import styles from '@patternfly/react-styles/css/components/Toolbar/toolbar.mjs';
import { css } from '@patternfly/react-styles';
import cssWidth from '@patternfly/react-tokens/dist/esm/c_toolbar__item_Width';
import { formatBreakpointMods, toCamel } from '../../helpers/util';
import { Divider } from '../Divider';
import { PageContext } from '../Page/PageContext';
export var ToolbarItemVariant;
(function (ToolbarItemVariant) {
    ToolbarItemVariant["separator"] = "separator";
    ToolbarItemVariant["bulk-select"] = "bulk-select";
    ToolbarItemVariant["overflow-menu"] = "overflow-menu";
    ToolbarItemVariant["pagination"] = "pagination";
    ToolbarItemVariant["search-filter"] = "search-filter";
    ToolbarItemVariant["label"] = "label";
    ToolbarItemVariant["chip-group"] = "chip-group";
    ToolbarItemVariant["expand-all"] = "expand-all";
})(ToolbarItemVariant || (ToolbarItemVariant = {}));
export const ToolbarItem = (_a) => {
    var { className, variant, visibility, spacer, widths, align, alignSelf, alignItems, id, children, isAllExpanded, isOverflowContainer } = _a, props = __rest(_a, ["className", "variant", "visibility", "spacer", "widths", "align", "alignSelf", "alignItems", "id", "children", "isAllExpanded", "isOverflowContainer"]);
    if (variant === ToolbarItemVariant.separator) {
        return React.createElement(Divider, Object.assign({ className: css(styles.modifiers.vertical, className) }, props));
    }
    const widthStyles = {};
    if (widths) {
        Object.entries(widths || {}).map(([breakpoint, value]) => (widthStyles[`${cssWidth.name}${breakpoint !== 'default' ? `-on-${breakpoint}` : ''}`] = value));
    }
    return (React.createElement(PageContext.Consumer, null, ({ width, getBreakpoint }) => (React.createElement("div", Object.assign({ className: css(styles.toolbarItem, variant &&
            styles.modifiers[toCamel(variant)], isAllExpanded && styles.modifiers.expanded, isOverflowContainer && styles.modifiers.overflowContainer, formatBreakpointMods(visibility, styles, '', getBreakpoint(width)), formatBreakpointMods(align, styles, '', getBreakpoint(width)), formatBreakpointMods(spacer, styles, '', getBreakpoint(width)), alignItems === 'start' && styles.modifiers.alignItemsStart, alignItems === 'center' && styles.modifiers.alignItemsCenter, alignItems === 'baseline' && styles.modifiers.alignItemsBaseline, alignSelf === 'start' && styles.modifiers.alignSelfStart, alignSelf === 'center' && styles.modifiers.alignSelfCenter, alignSelf === 'baseline' && styles.modifiers.alignSelfBaseline, className) }, (variant === 'label' && { 'aria-hidden': true }), { id: id }, props, (widths && { style: Object.assign(Object.assign({}, widthStyles), props.style) })), children))));
};
ToolbarItem.displayName = 'ToolbarItem';
//# sourceMappingURL=ToolbarItem.js.map