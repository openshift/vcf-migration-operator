import { __rest } from "tslib";
import * as React from 'react';
import { css } from '@patternfly/react-styles';
import styles from '@patternfly/react-styles/css/components/AppLauncher/app-launcher.mjs';
export const ApplicationLauncherText = (_a) => {
    var { className = '', children } = _a, props = __rest(_a, ["className", "children"]);
    return (React.createElement("span", Object.assign({ className: css(`${styles.appLauncherMenuItem}-text`, className) }, props), children));
};
ApplicationLauncherText.displayName = 'ApplicationLauncherText';
//# sourceMappingURL=ApplicationLauncherText.js.map