import { __rest } from "tslib";
import * as React from 'react';
import { css } from '@patternfly/react-styles';
import styles from '@patternfly/react-styles/css/components/Sidebar/sidebar.mjs';
export const SidebarContent = (_a) => {
    var { className, children, hasNoBackground, hasPadding } = _a, props = __rest(_a, ["className", "children", "hasNoBackground", "hasPadding"]);
    return (React.createElement("div", Object.assign({ className: css(styles.sidebarContent, hasNoBackground && styles.modifiers.noBackground, hasPadding && styles.modifiers.padding, className) }, props), children));
};
SidebarContent.displayName = 'SidebarContent';
//# sourceMappingURL=SidebarContent.js.map