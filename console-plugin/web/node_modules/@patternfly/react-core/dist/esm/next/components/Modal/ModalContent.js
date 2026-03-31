import { __rest } from "tslib";
import * as React from 'react';
import { FocusTrap } from '../../../helpers';
import bullsEyeStyles from '@patternfly/react-styles/css/layouts/Bullseye/bullseye.mjs';
import { css } from '@patternfly/react-styles';
import { getOUIAProps } from '../../../helpers';
import { Backdrop } from '../../../components/Backdrop';
import { ModalBoxCloseButton } from './ModalBoxCloseButton';
import { ModalBox } from './ModalBox';
export const ModalContent = (_a) => {
    var { children, className, isOpen = false, 'aria-label': ariaLabel, 'aria-describedby': ariaDescribedby, 'aria-labelledby': ariaLabelledby, onClose, variant = 'default', position, positionOffset, width, maxWidth, boxId, backdropId, disableFocusTrap = false, ouiaId, ouiaSafe = true, elementToFocus } = _a, props = __rest(_a, ["children", "className", "isOpen", 'aria-label', 'aria-describedby', 'aria-labelledby', "onClose", "variant", "position", "positionOffset", "width", "maxWidth", "boxId", "backdropId", "disableFocusTrap", "ouiaId", "ouiaSafe", "elementToFocus"]);
    if (!isOpen) {
        return null;
    }
    const getAriaLabelledBy = () => {
        if (ariaLabelledby) {
            return ariaLabelledby;
        }
        if (ariaLabel) {
            return undefined;
        }
        return boxId;
    };
    const modalBox = (React.createElement(ModalBox, Object.assign({ className: css(className), variant: variant, position: position, positionOffset: positionOffset, "aria-label": ariaLabel, "aria-labelledby": getAriaLabelledBy(), "aria-describedby": ariaDescribedby }, getOUIAProps(ModalContent.displayName, ouiaId, ouiaSafe), { style: Object.assign(Object.assign({}, (width && { '--pf-v5-c-modal-box--Width': typeof width !== 'number' ? width : `${width}px` })), (maxWidth && {
            '--pf-v5-c-modal-box--MaxWidth': typeof maxWidth !== 'number' ? maxWidth : `${maxWidth}px`
        })) }, props, { id: boxId }),
        onClose && React.createElement(ModalBoxCloseButton, { onClose: (event) => onClose(event), ouiaId: ouiaId }),
        children));
    return (React.createElement(Backdrop, { id: backdropId },
        React.createElement(FocusTrap, { active: !disableFocusTrap, focusTrapOptions: {
                clickOutsideDeactivates: true,
                tabbableOptions: { displayCheck: 'none' },
                // FocusTrap's initialFocus can accept false as a value to prevent initial focus.
                // We want to prevent this in case false is ever passed in.
                initialFocus: elementToFocus || undefined
            }, className: css(bullsEyeStyles.bullseye) }, modalBox)));
};
ModalContent.displayName = 'ModalContent';
//# sourceMappingURL=ModalContent.js.map