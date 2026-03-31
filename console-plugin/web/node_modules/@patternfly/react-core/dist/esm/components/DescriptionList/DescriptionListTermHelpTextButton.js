import { __rest } from "tslib";
import * as React from 'react';
import styles from '@patternfly/react-styles/css/components/DescriptionList/description-list.mjs';
import { css } from '@patternfly/react-styles';
export const DescriptionListTermHelpTextButton = (_a) => {
    var { children, className } = _a, props = __rest(_a, ["children", "className"]);
    const helpTextRef = React.createRef();
    const handleKeys = (event) => {
        if (!helpTextRef.current || helpTextRef.current !== event.target) {
            return;
        }
        const key = event.key;
        if (key === 'Enter' || key === ' ') {
            event.preventDefault();
            helpTextRef.current.click();
        }
    };
    return (React.createElement("span", Object.assign({ ref: helpTextRef, className: css(className, styles.descriptionListText, styles.modifiers.helpText), role: "button", type: "button", tabIndex: 0, onKeyDown: (event) => handleKeys(event) }, props), children));
};
DescriptionListTermHelpTextButton.displayName = 'DescriptionListTermHelpTextButton';
//# sourceMappingURL=DescriptionListTermHelpTextButton.js.map