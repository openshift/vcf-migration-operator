import * as React from 'react';
import { css } from '@patternfly/react-styles';
import styles from '@patternfly/react-styles/css/components/Wizard/wizard.mjs';
import { Button, ButtonVariant } from '../../../components/Button';
export const WizardFooterInternal = ({ onNext, onBack, onClose, isValid, firstStep, activeStep, nextButtonText, backButtonText, cancelButtonText }) => (React.createElement("footer", { className: css(styles.wizardFooter) },
    !activeStep.hideBackButton && (React.createElement(Button, { variant: ButtonVariant.secondary, onClick: onBack, isDisabled: firstStep }, backButtonText)),
    React.createElement(Button, { variant: ButtonVariant.primary, type: "submit", onClick: onNext, isDisabled: !isValid }, nextButtonText),
    !activeStep.hideCancelButton && (React.createElement("div", { className: styles.wizardFooterCancel },
        React.createElement(Button, { variant: ButtonVariant.link, onClick: onClose }, cancelButtonText)))));
WizardFooterInternal.displayName = 'WizardFooterInternal';
//# sourceMappingURL=WizardFooterInternal.js.map