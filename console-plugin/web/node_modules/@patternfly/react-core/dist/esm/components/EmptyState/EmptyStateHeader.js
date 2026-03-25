import { __rest } from "tslib";
import * as React from 'react';
import { css } from '@patternfly/react-styles';
import styles from '@patternfly/react-styles/css/components/EmptyState/empty-state.mjs';
export const EmptyStateHeader = (_a) => {
    var { children, className, titleClassName, titleText, headingLevel: HeadingLevel = 'h1', icon } = _a, props = __rest(_a, ["children", "className", "titleClassName", "titleText", "headingLevel", "icon"]);
    return (React.createElement("div", Object.assign({ className: css(`${styles.emptyState}__header`, className) }, props),
        icon,
        (titleText || children) && (React.createElement("div", { className: css(`${styles.emptyState}__title`) },
            titleText && (React.createElement(HeadingLevel, { className: css(styles.emptyStateTitleText, titleClassName) }, titleText)),
            children))));
};
EmptyStateHeader.displayName = 'EmptyStateHeader';
//# sourceMappingURL=EmptyStateHeader.js.map