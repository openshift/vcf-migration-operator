import { __rest } from "tslib";
import * as React from 'react';
import { css } from '@patternfly/react-styles';
import styles from '@patternfly/react-styles/css/components/EmptyState/empty-state.mjs';
export var EmptyStateVariant;
(function (EmptyStateVariant) {
    EmptyStateVariant["xs"] = "xs";
    EmptyStateVariant["sm"] = "sm";
    EmptyStateVariant["lg"] = "lg";
    EmptyStateVariant["xl"] = "xl";
    EmptyStateVariant["full"] = "full";
})(EmptyStateVariant || (EmptyStateVariant = {}));
export const EmptyState = (_a) => {
    var { children, className, variant = EmptyStateVariant.full, isFullHeight } = _a, props = __rest(_a, ["children", "className", "variant", "isFullHeight"]);
    return (React.createElement("div", Object.assign({ className: css(styles.emptyState, variant === 'xs' && styles.modifiers.xs, variant === 'sm' && styles.modifiers.sm, variant === 'lg' && styles.modifiers.lg, variant === 'xl' && styles.modifiers.xl, isFullHeight && styles.modifiers.fullHeight, className) }, props),
        React.createElement("div", { className: css(styles.emptyStateContent) }, children)));
};
EmptyState.displayName = 'EmptyState';
//# sourceMappingURL=EmptyState.js.map