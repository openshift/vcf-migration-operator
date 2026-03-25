import { __rest } from "tslib";
import * as React from 'react';
import styles from '@patternfly/react-styles/css/components/Nav/nav.mjs';
import { css } from '@patternfly/react-styles';
import AngleLeftIcon from '@patternfly/react-icons/dist/esm/icons/angle-left-icon';
import AngleRightIcon from '@patternfly/react-icons/dist/esm/icons/angle-right-icon';
import { getLanguageDirection, isElementInView } from '../../helpers/util';
import { NavContext } from './Nav';
import { PageSidebarContext } from '../Page/PageSidebar';
import { getResizeObserver } from '../../helpers/resizeObserver';
class NavList extends React.Component {
    constructor() {
        super(...arguments);
        this.direction = 'ltr';
        this.state = {
            scrollViewAtStart: false,
            scrollViewAtEnd: false
        };
        this.navList = React.createRef();
        this.observer = () => { };
        this.handleScrollButtons = () => {
            const container = this.navList.current;
            if (container) {
                // check if it elements are in view
                const scrollViewAtStart = isElementInView(container, container.firstChild, false);
                const scrollViewAtEnd = isElementInView(container, container.lastChild, false);
                this.setState({
                    scrollViewAtStart,
                    scrollViewAtEnd
                });
                this.context.updateIsScrollable(!scrollViewAtStart || !scrollViewAtEnd);
            }
        };
        this.scrollBack = () => {
            // find first Element that is fully in view on the left, then scroll to the element before it
            const container = this.navList.current;
            if (container) {
                const childrenArr = Array.from(container.children);
                let firstElementInView;
                let lastElementOutOfView;
                for (let i = 0; i < childrenArr.length && !firstElementInView; i++) {
                    if (isElementInView(container, childrenArr[i], false)) {
                        firstElementInView = childrenArr[i];
                        lastElementOutOfView = childrenArr[i - 1];
                    }
                }
                if (lastElementOutOfView) {
                    if (this.direction === 'ltr') {
                        // LTR scrolls left to go back
                        container.scrollLeft -= lastElementOutOfView.scrollWidth;
                    }
                    else {
                        // RTL scrolls right to go back
                        container.scrollLeft += lastElementOutOfView.scrollWidth;
                    }
                }
                this.handleScrollButtons();
            }
        };
        this.scrollForward = () => {
            // find last Element that is fully in view on the right, then scroll to the element after it
            const container = this.navList.current;
            if (container) {
                const childrenArr = Array.from(container.children);
                let lastElementInView;
                let firstElementOutOfView;
                for (let i = childrenArr.length - 1; i >= 0 && !lastElementInView; i--) {
                    if (isElementInView(container, childrenArr[i], false)) {
                        lastElementInView = childrenArr[i];
                        firstElementOutOfView = childrenArr[i + 1];
                    }
                }
                if (firstElementOutOfView) {
                    if (this.direction === 'ltr') {
                        // LTR scrolls right to go forward
                        container.scrollLeft += firstElementOutOfView.scrollWidth;
                    }
                    else {
                        // RTL scrolls left to go forward
                        container.scrollLeft -= firstElementOutOfView.scrollWidth;
                    }
                }
                this.handleScrollButtons();
            }
        };
    }
    componentDidMount() {
        this.observer = getResizeObserver(this.navList.current, this.handleScrollButtons, true);
        this.direction = getLanguageDirection(this.navList.current);
        this.handleScrollButtons();
    }
    componentWillUnmount() {
        this.observer();
    }
    componentDidUpdate() {
        this.direction = getLanguageDirection(this.navList.current);
    }
    render() {
        const _a = this.props, { children, className, ariaLeftScroll, ariaRightScroll, backScrollAriaLabel, forwardScrollAriaLabel } = _a, props = __rest(_a, ["children", "className", "ariaLeftScroll", "ariaRightScroll", "backScrollAriaLabel", "forwardScrollAriaLabel"]);
        const { scrollViewAtStart, scrollViewAtEnd } = this.state;
        return (React.createElement(NavContext.Consumer, null, ({ isHorizontal }) => (React.createElement(PageSidebarContext.Consumer, null, ({ isSidebarOpen }) => (React.createElement(React.Fragment, null,
            isHorizontal && (React.createElement("button", { className: css(styles.navScrollButton), "aria-label": backScrollAriaLabel || ariaLeftScroll, onClick: this.scrollBack, disabled: scrollViewAtStart, tabIndex: isSidebarOpen ? null : -1 },
                React.createElement(AngleLeftIcon, null))),
            React.createElement("ul", Object.assign({ ref: this.navList, className: css(styles.navList, className), onScroll: this.handleScrollButtons, role: "list" }, props), children),
            isHorizontal && (React.createElement("button", { className: css(styles.navScrollButton), "aria-label": forwardScrollAriaLabel || ariaRightScroll, onClick: this.scrollForward, disabled: scrollViewAtEnd, tabIndex: isSidebarOpen ? null : -1 },
                React.createElement(AngleRightIcon, null)))))))));
    }
}
NavList.displayName = 'NavList';
NavList.contextType = NavContext;
NavList.defaultProps = {
    ariaLeftScroll: 'Scroll left',
    backScrollAriaLabel: 'Scroll back',
    ariaRightScroll: 'Scroll right',
    forwardScrollAriaLabel: 'Scroll foward'
};
export { NavList };
//# sourceMappingURL=NavList.js.map