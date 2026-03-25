import { __rest } from "tslib";
import * as React from 'react';
import { css } from '@patternfly/react-styles';
import styles from '@patternfly/react-styles/css/components/EmptyState/empty-state.mjs';
import { Spinner } from '../Spinner';
import cssIconColor from '@patternfly/react-tokens/dist/esm/c_empty_state__icon_Color';
const isSpinner = (icon) => icon.type === Spinner;
export const EmptyStateIcon = (_a) => {
    var { className, icon: IconComponent, color } = _a, props = __rest(_a, ["className", "icon", "color"]);
    const iconIsSpinner = isSpinner(React.createElement(IconComponent, null));
    return (React.createElement("div", Object.assign({ className: css(styles.emptyStateIcon) }, (color && !iconIsSpinner && { style: { [cssIconColor.name]: color } })),
        React.createElement(IconComponent, Object.assign({ className: className, "aria-hidden": !iconIsSpinner }, props))));
};
EmptyStateIcon.displayName = 'EmptyStateIcon';
//# sourceMappingURL=EmptyStateIcon.js.map