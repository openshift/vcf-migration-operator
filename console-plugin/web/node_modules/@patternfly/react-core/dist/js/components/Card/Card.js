"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.Card = exports.CardContext = void 0;
const tslib_1 = require("tslib");
const React = tslib_1.__importStar(require("react"));
const card_1 = tslib_1.__importDefault(require("@patternfly/react-styles/css/components/Card/card"));
const react_styles_1 = require("@patternfly/react-styles");
const helpers_1 = require("../../helpers");
exports.CardContext = React.createContext({
    cardId: '',
    registerTitleId: () => { },
    isExpanded: false,
    isClickable: false,
    isSelectable: false,
    isSelected: false,
    isClicked: false,
    isDisabled: false
});
const Card = (_a) => {
    var { children, id = '', className, component = 'div', isCompact = false, isSelectable = false, isClickable = false, isDisabled = false, isSelectableRaised = false, isSelected = false, isClicked = false, isDisabledRaised = false, isFlat = false, isExpanded = false, isRounded = false, isLarge = false, isFullHeight = false, isPlain = false, ouiaId, ouiaSafe = true, hasSelectableInput = false, selectableInputAriaLabel, onSelectableInputChange = () => { } } = _a, props = tslib_1.__rest(_a, ["children", "id", "className", "component", "isCompact", "isSelectable", "isClickable", "isDisabled", "isSelectableRaised", "isSelected", "isClicked", "isDisabledRaised", "isFlat", "isExpanded", "isRounded", "isLarge", "isFullHeight", "isPlain", "ouiaId", "ouiaSafe", "hasSelectableInput", "selectableInputAriaLabel", "onSelectableInputChange"]);
    const Component = component;
    const ouiaProps = (0, helpers_1.useOUIAProps)(exports.Card.displayName, ouiaId, ouiaSafe);
    const [titleId, setTitleId] = React.useState('');
    const [ariaProps, setAriaProps] = React.useState();
    if (isCompact && isLarge) {
        // eslint-disable-next-line no-console
        console.warn('Card: Cannot use isCompact with isLarge. Defaulting to isCompact');
        isLarge = false;
    }
    const getSelectableModifiers = () => {
        if (isDisabledRaised) {
            return (0, react_styles_1.css)(card_1.default.modifiers.nonSelectableRaised);
        }
        if (isSelectableRaised) {
            return (0, react_styles_1.css)(card_1.default.modifiers.selectableRaised, isSelected && card_1.default.modifiers.selectedRaised);
        }
        if (isSelectable && isClickable) {
            return (0, react_styles_1.css)(card_1.default.modifiers.selectable, card_1.default.modifiers.clickable, (isSelected || isClicked) && card_1.default.modifiers.current);
        }
        if (isSelectable) {
            return (0, react_styles_1.css)(card_1.default.modifiers.selectable, isSelected && card_1.default.modifiers.selected);
        }
        if (isClickable) {
            return (0, react_styles_1.css)(card_1.default.modifiers.clickable, isClicked && card_1.default.modifiers.current);
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
    return (React.createElement(exports.CardContext.Provider, { value: {
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
        React.createElement(Component, Object.assign({ id: id, className: (0, react_styles_1.css)(card_1.default.card, isCompact && card_1.default.modifiers.compact, isExpanded && card_1.default.modifiers.expanded, isFlat && card_1.default.modifiers.flat, isRounded && card_1.default.modifiers.rounded, isLarge && card_1.default.modifiers.displayLg, isFullHeight && card_1.default.modifiers.fullHeight, isPlain && card_1.default.modifiers.plain, getSelectableModifiers(), isDisabled && card_1.default.modifiers.disabled, className), tabIndex: isSelectableRaised ? '0' : undefined }, props, ouiaProps), children)));
};
exports.Card = Card;
exports.Card.displayName = 'Card';
//# sourceMappingURL=Card.js.map