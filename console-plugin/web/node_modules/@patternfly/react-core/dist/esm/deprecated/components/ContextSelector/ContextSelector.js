import { __rest } from "tslib";
import * as React from 'react';
import styles from '@patternfly/react-styles/css/components/ContextSelector/context-selector.mjs';
import { css } from '@patternfly/react-styles';
import { ContextSelectorToggle } from './ContextSelectorToggle';
import { ContextSelectorMenuList } from './ContextSelectorMenuList';
import { ContextSelectorContext } from './contextSelectorConstants';
import { FocusTrap, getUniqueId } from '../../../helpers';
import { Popper } from '../../../helpers/Popper/Popper';
import { getOUIAProps, getDefaultOUIAId } from '../../../helpers';
import { SearchInput } from '../../../components/SearchInput';
class ContextSelector extends React.Component {
    constructor(props) {
        super(props);
        this.parentRef = React.createRef();
        this.popperRef = React.createRef();
        this.state = {
            ouiaStateId: getDefaultOUIAId(ContextSelector.displayName)
        };
    }
    render() {
        const _a = this.props, { children, className, isOpen, isFullHeight, onToggle, onSelect, screenReaderLabel, toggleText, searchButtonAriaLabel, searchInputValue, onSearchInputChange, searchInputPlaceholder, onSearchButtonClick, menuAppendTo, ouiaId, ouiaSafe, isPlain, isText, footer, disableFocusTrap, isFlipEnabled, id, zIndex, menuAriaLabel } = _a, props = __rest(_a, ["children", "className", "isOpen", "isFullHeight", "onToggle", "onSelect", "screenReaderLabel", "toggleText", "searchButtonAriaLabel", "searchInputValue", "onSearchInputChange", "searchInputPlaceholder", "onSearchButtonClick", "menuAppendTo", "ouiaId", "ouiaSafe", "isPlain", "isText", "footer", "disableFocusTrap", "isFlipEnabled", "id", "zIndex", "menuAriaLabel"]);
        const uniqueId = id || getUniqueId();
        const toggleId = `pf-v5-context-selector-toggle-id-${uniqueId}`;
        const screenReaderLabelId = `pf-v5-context-selector-label-id-${uniqueId}`;
        const isStatic = isFlipEnabled && menuAppendTo !== 'inline';
        const menuContainer = (React.createElement("div", { className: css(styles.contextSelectorMenu, isStatic && styles.modifiers.static), id: uniqueId }, isOpen && (React.createElement(FocusTrap, { active: !disableFocusTrap, focusTrapOptions: { clickOutsideDeactivates: true, tabbableOptions: { displayCheck: 'none' } } },
            React.createElement("div", { className: css(styles.contextSelectorMenuSearch) },
                React.createElement(SearchInput, { submitSearchButtonLabel: searchButtonAriaLabel, placeholder: searchInputPlaceholder, onChange: onSearchInputChange, onSearch: (event, _value, _) => {
                        onSearchButtonClick(event);
                    }, value: searchInputValue })),
            React.createElement(ContextSelectorContext.Provider, { value: { onSelect } },
                React.createElement(ContextSelectorMenuList, { isOpen: isOpen, "aria-label": menuAriaLabel }, children)),
            footer))));
        const popperContainer = (React.createElement("div", Object.assign({ className: css(styles.contextSelector, isOpen && styles.modifiers.expanded, className), ref: this.popperRef }, props), isOpen && menuContainer));
        const mainContainer = (React.createElement("div", Object.assign({ className: css(styles.contextSelector, isOpen && styles.modifiers.expanded, isFullHeight && styles.modifiers.fullHeight, className), ref: this.parentRef }, getOUIAProps(ContextSelector.displayName, ouiaId !== undefined ? ouiaId : this.state.ouiaStateId, ouiaSafe), props),
            screenReaderLabel && (React.createElement("span", { id: screenReaderLabelId, hidden: true }, screenReaderLabel)),
            React.createElement(ContextSelectorToggle, { onToggle: onToggle, isOpen: isOpen, toggleText: toggleText, id: toggleId, parentRef: menuAppendTo === 'inline' ? this.parentRef : this.popperRef, "aria-labelledby": `${screenReaderLabelId} ${toggleId}`, isPlain: isPlain, isText: isText }),
            isOpen && menuAppendTo === 'inline' && menuContainer));
        const getParentElement = () => {
            if (this.parentRef && this.parentRef.current) {
                return this.parentRef.current.parentElement;
            }
            return null;
        };
        return menuAppendTo === 'inline' ? (mainContainer) : (React.createElement(Popper, { trigger: mainContainer, triggerRef: this.parentRef, popper: popperContainer, popperRef: this.popperRef, appendTo: menuAppendTo === 'parent' ? getParentElement() : menuAppendTo, isVisible: isOpen, zIndex: zIndex }));
    }
}
ContextSelector.displayName = 'ContextSelector';
ContextSelector.defaultProps = {
    children: null,
    className: '',
    isOpen: false,
    onToggle: () => undefined,
    onSelect: () => undefined,
    screenReaderLabel: '',
    toggleText: '',
    searchButtonAriaLabel: 'Search menu items',
    searchInputValue: '',
    onSearchInputChange: () => undefined,
    searchInputPlaceholder: 'Search',
    onSearchButtonClick: () => undefined,
    menuAppendTo: 'inline',
    ouiaSafe: true,
    disableFocusTrap: false,
    footer: null,
    isPlain: false,
    isText: false,
    isFlipEnabled: true,
    zIndex: 9999
};
export { ContextSelector };
//# sourceMappingURL=ContextSelector.js.map