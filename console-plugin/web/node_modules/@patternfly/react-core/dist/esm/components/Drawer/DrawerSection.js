import { __rest } from "tslib";
import * as React from 'react';
import styles from '@patternfly/react-styles/css/components/Drawer/drawer.mjs';
import { css } from '@patternfly/react-styles';
import { DrawerColorVariant } from './Drawer';
export const DrawerSection = (_a) => {
    var { className = '', children, colorVariant = DrawerColorVariant.default } = _a, props = __rest(_a, ["className", "children", "colorVariant"]);
    return (React.createElement("div", Object.assign({ className: css(styles.drawerSection, colorVariant === DrawerColorVariant.light200 && styles.modifiers.light_200, colorVariant === DrawerColorVariant.noBackground && styles.modifiers.noBackground, className) }, props), children));
};
DrawerSection.displayName = 'DrawerSection';
//# sourceMappingURL=DrawerSection.js.map