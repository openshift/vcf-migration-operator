import { __rest } from "tslib";
import * as React from 'react';
import styles from '@patternfly/react-styles/css/components/Card/card.mjs';
import { css } from '@patternfly/react-styles';
import { useOUIAProps } from '../../helpers';
export const CardContext = React.createContext({
    cardId: '',
    registerTitleId: () => { },
    isExpanded: false,
    isClickable: false,
    isSelectable: false,
    isSelected: false,
    isClicked: false,
    isDisabled: false
});
export const Card = (_a) => {
    var { children, id = '', className, component = 'div', isCompact = false, isSelectable = false, isClickable = false, isDisabled = false, isSelectableRaised = false, isSelected = false, isClicked = false, isDisabledRaised = false, isFlat = false, isExpanded = false, isRounded = false, isLarge = false, isFullHeight = false, isPlain = false, ouiaId, ouiaSafe = true, hasSelectableInput = false, selectableInputAriaLabel, onSelectableInputChange = () => { } } = _a, props = __rest(_a, ["children", "id", "className", "component", "isCompact", "isSelectable", "isClickable", "isDisabled", "isSelectableRaised", "isSelected", "isClicked", "isDisabledRaised", "isFlat", "isExpanded", "isRounded", "isLarge", "isFullHeight", "isPlain", "ouiaId", "ouiaSafe", "hasSelectableInput", "selectableInputAriaLabel", "onSelectableInputChange"]);
    const Component = component;
    const ouiaProps = useOUIAProps(Card.displayName, ouiaId, ouiaSafe);
    const [titleId, setTitleId] = React.useState('');
    const [ariaProps, setAriaProps] = React.useState();
    if (isCompact && isLarge) {
        // eslint-disable-next-line no-console
        console.warn('Card: Cannot use isCompact with isLarge. Defaulting to isCompact');
        isLarge = false;
    }
    const getSelectableModifiers = () => {
        if (isDisabledRaised) {
            return css(styles.modifiers.nonSelectableRaised);
        }
        if (isSelectableRaised) {
            return css(styles.modifiers.selectableRaised, isSelected && styles.modifiers.selectedRaised);
        }
        if (isSelectable && isClickable) {
            return css(styles.modifiers.selectable, styles.modifiers.clickable, (isSelected || isClicked) && styles.modifiers.current);
        }
        if (isSelectable) {
            return css(styles.modifiers.selectable, isSelected && styles.modifiers.selected);
        }
        if (isClickable) {
            return css(styles.modifiers.clickable, isClicked && styles.modifiers.current);
        }
        return '';
    };
    const containsCardTitleChildRef = React.useRef(false);
    const registerTitleId = (id) => {
        setTitleId(id);
        containsCardTitleChildRef.current = !!id;
    };
    React.useEffect(() => {
        if (selectableInputAriaLabel) {
            setAriaProps({ 'aria-label': selectableInputAriaLabel });
        }
        else if (titleId) {
            setAriaProps({ 'aria-labelledby': titleId });
        }
        else if (hasSelectableInput && !containsCardTitleChildRef.current) {
            setAriaProps({});
            // eslint-disable-next-line no-console
            console.warn('If no CardTitle component is passed as a child of Card the selectableInputAriaLabel prop must be passed');
        }
    }, [hasSelectableInput, selectableInputAriaLabel, titleId]);
    return (React.createElement(CardContext.Provider, { value: {
            cardId: id,
            registerTitleId,
            isExpanded,
            isClickable,
            isSelectable,
            isSelected,
            isClicked,
            isDisabled,
            // TODO: Remove hasSelectableInput when deprecated prop is removed
            hasSelectableInput
        } },
        hasSelectableInput && (React.createElement("input", Object.assign({ className: "pf-v5-screen-reader", id: `${id}-input` }, ariaProps, { type: "checkbox", checked: isSelected, onChange: (event) => onSelectableInputChange(event, id), disabled: isDisabledRaised, tabIndex: -1 }))),
        React.createElement(Component, Object.assign({ id: id, className: css(styles.card, isCompact && styles.modifiers.compact, isExpanded && styles.modifiers.expanded, isFlat && styles.modifiers.flat, isRounded && styles.modifiers.rounded, isLarge && styles.modifiers.displayLg, isFullHeight && styles.modifiers.fullHeight, isPlain && styles.modifiers.plain, getSelectableModifiers(), isDisabled && styles.modifiers.disabled, className), tabIndex: isSelectableRaised ? '0' : undefined }, props, ouiaProps), children)));
};
Card.displayName = 'Card';
//# sourceMappingURL=Card.js.map