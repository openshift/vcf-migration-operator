import { __rest } from "tslib";
import * as React from 'react';
import { useState } from 'react';
import styles from '@patternfly/react-styles/css/components/Label/label.mjs';
import labelGrpStyles from '@patternfly/react-styles/css/components/Label/label-group.mjs';
import { Button } from '../Button';
import { Tooltip } from '../Tooltip';
import { css } from '@patternfly/react-styles';
import TimesIcon from '@patternfly/react-icons/dist/esm/icons/times-icon';
import { useIsomorphicLayoutEffect } from '../../helpers';
import cssTextMaxWidth from '@patternfly/react-tokens/dist/esm/c_label__text_MaxWidth';
const colorStyles = {
    blue: styles.modifiers.blue,
    cyan: styles.modifiers.cyan,
    green: styles.modifiers.green,
    orange: styles.modifiers.orange,
    purple: styles.modifiers.purple,
    red: styles.modifiers.red,
    gold: styles.modifiers.gold,
    grey: ''
};
export const Label = (_a) => {
    var { children, className = '', color = 'grey', variant = 'filled', isCompact = false, isDisabled = false, isEditable = false, editableProps, textMaxWidth, tooltipPosition, icon, onClose, onClick: onLabelClick, onEditCancel, onEditComplete, closeBtn, closeBtnAriaLabel, closeBtnProps, href, isOverflowLabel, render } = _a, props = __rest(_a, ["children", "className", "color", "variant", "isCompact", "isDisabled", "isEditable", "editableProps", "textMaxWidth", "tooltipPosition", "icon", "onClose", "onClick", "onEditCancel", "onEditComplete", "closeBtn", "closeBtnAriaLabel", "closeBtnProps", "href", "isOverflowLabel", "render"]);
    const [isEditableActive, setIsEditableActive] = useState(false);
    const [currValue, setCurrValue] = useState(children);
    const editableButtonRef = React.useRef();
    const editableInputRef = React.useRef();
    React.useEffect(() => {
        document.addEventListener('mousedown', onDocMouseDown);
        document.addEventListener('keydown', onKeyDown);
        return () => {
            document.removeEventListener('mousedown', onDocMouseDown);
            document.removeEventListener('keydown', onKeyDown);
        };
    });
    React.useEffect(() => {
        if (onLabelClick && href) {
            // eslint-disable-next-line no-console
            console.warn('Link labels cannot have onClick passed, this results in invalid HTML. Please remove either the href or onClick prop.');
        }
        else if (onLabelClick && isEditable) {
            // eslint-disable-next-line no-console
            console.warn('Editable labels cannot have onClick passed, clicking starts the label edit process. Please remove either the isEditable or onClick prop.');
        }
    }, [onLabelClick, href, isEditable]);
    const onDocMouseDown = (event) => {
        if (isEditableActive &&
            editableInputRef &&
            editableInputRef.current &&
            !editableInputRef.current.contains(event.target)) {
            if (editableInputRef.current.value) {
                onEditComplete && onEditComplete(event, editableInputRef.current.value);
            }
            setIsEditableActive(false);
        }
    };
    const onKeyDown = (event) => {
        var _a, _b;
        const key = event.key;
        if ((!isEditableActive &&
            (!editableButtonRef ||
                !editableButtonRef.current ||
                !editableButtonRef.current.contains(event.target))) ||
            (isEditableActive &&
                (!editableInputRef || !editableInputRef.current || !editableInputRef.current.contains(event.target)))) {
            return;
        }
        if (isEditableActive && (key === 'Enter' || key === 'Tab')) {
            event.preventDefault();
            event.stopImmediatePropagation();
            if (editableInputRef.current.value) {
                onEditComplete && onEditComplete(event, editableInputRef.current.value);
            }
            setIsEditableActive(false);
            (_a = editableButtonRef === null || editableButtonRef === void 0 ? void 0 : editableButtonRef.current) === null || _a === void 0 ? void 0 : _a.focus();
        }
        if (isEditableActive && key === 'Escape') {
            event.preventDefault();
            event.stopImmediatePropagation();
            // Reset div text to initial children prop - pre-edit
            if (editableInputRef.current.value) {
                editableInputRef.current.value = children;
                onEditCancel && onEditCancel(event, children);
            }
            setIsEditableActive(false);
            (_b = editableButtonRef === null || editableButtonRef === void 0 ? void 0 : editableButtonRef.current) === null || _b === void 0 ? void 0 : _b.focus();
        }
        if (!isEditableActive && key === 'Enter') {
            event.preventDefault();
            event.stopImmediatePropagation();
            setIsEditableActive(true);
            // Set cursor position to end of text
            const el = event.target;
            const range = document.createRange();
            const sel = window.getSelection();
            range.selectNodeContents(el);
            range.collapse(false);
            sel.removeAllRanges();
            sel.addRange(range);
        }
    };
    const isClickableDisabled = (href || onLabelClick) && isDisabled;
    const defaultCloseButton = (React.createElement(Button, Object.assign({ type: "button", variant: "plain", onClick: onClose, "aria-label": closeBtnAriaLabel || `Close ${children}` }, (isClickableDisabled && { isDisabled: true }), closeBtnProps),
        React.createElement(TimesIcon, null)));
    const closeButton = React.createElement("span", { className: css(styles.labelActions) }, closeBtn || defaultCloseButton);
    const textRef = React.createRef();
    // ref to apply tooltip when rendered is used
    const componentRef = React.useRef();
    const [isTooltipVisible, setIsTooltipVisible] = React.useState(false);
    useIsomorphicLayoutEffect(() => {
        const currTextRef = isEditable ? editableButtonRef : textRef;
        if (!isEditableActive) {
            setIsTooltipVisible(currTextRef.current && currTextRef.current.offsetWidth < currTextRef.current.scrollWidth);
        }
    }, [isEditableActive]);
    const content = (React.createElement(React.Fragment, null,
        icon && React.createElement("span", { className: css(styles.labelIcon) }, icon),
        React.createElement("span", Object.assign({ ref: textRef, className: css(styles.labelText) }, (textMaxWidth && {
            style: {
                [cssTextMaxWidth.name]: textMaxWidth
            }
        })), children)));
    React.useEffect(() => {
        if (isEditableActive && editableInputRef) {
            editableInputRef.current && editableInputRef.current.focus();
        }
    }, [editableInputRef, isEditableActive]);
    const updateVal = () => {
        setCurrValue(editableInputRef.current.value);
    };
    let LabelComponentChildElement = 'span';
    if (href) {
        LabelComponentChildElement = 'a';
    }
    else if (isEditable || (onLabelClick && !isOverflowLabel)) {
        LabelComponentChildElement = 'button';
    }
    const clickableLabelProps = {
        type: 'button',
        onClick: onLabelClick
    };
    const isButton = LabelComponentChildElement === 'button';
    const labelComponentChildProps = Object.assign(Object.assign(Object.assign(Object.assign(Object.assign(Object.assign(Object.assign({ className: css(styles.labelContent) }, (isTooltipVisible && { tabIndex: 0 })), (href && { href })), (href && isDisabled && { onClick: (event) => event.preventDefault() })), (isButton && clickableLabelProps)), (isEditable && Object.assign({ ref: editableButtonRef, onClick: (e) => {
            setIsEditableActive(true);
            e.stopPropagation();
        } }, editableProps))), (isClickableDisabled && isButton && { disabled: true })), (isClickableDisabled && href && { tabIndex: -1, 'aria-disabled': true }));
    let labelComponentChild = (React.createElement(LabelComponentChildElement, Object.assign({}, labelComponentChildProps), content));
    if (render) {
        labelComponentChild = (React.createElement(React.Fragment, null,
            isTooltipVisible && React.createElement(Tooltip, { triggerRef: componentRef, content: children, position: tooltipPosition }),
            render({
                className: styles.labelContent,
                content,
                componentRef
            })));
    }
    else if (isTooltipVisible) {
        labelComponentChild = (React.createElement(Tooltip, { content: children, position: tooltipPosition }, labelComponentChild));
    }
    const LabelComponent = (isOverflowLabel ? 'button' : 'span');
    return (React.createElement(LabelComponent, Object.assign({}, props, { className: css(styles.label, isClickableDisabled && styles.modifiers.disabled, colorStyles[color], variant === 'outline' && styles.modifiers.outline, isOverflowLabel && styles.modifiers.overflow, isCompact && styles.modifiers.compact, isEditable && labelGrpStyles.modifiers.editable, isEditableActive && styles.modifiers.editableActive, className), onClick: isOverflowLabel ? onLabelClick : undefined }, (LabelComponent === 'button' && { type: 'button' })),
        !isEditableActive && labelComponentChild,
        !isEditableActive && onClose && closeButton,
        isEditableActive && (React.createElement("input", Object.assign({ className: css(styles.labelContent), type: "text", id: "editable-input", ref: editableInputRef, value: currValue, onChange: updateVal }, editableProps)))));
};
Label.displayName = 'Label';
//# sourceMappingURL=Label.js.map