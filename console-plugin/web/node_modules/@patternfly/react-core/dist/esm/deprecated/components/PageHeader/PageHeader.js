import { __rest } from "tslib";
import * as React from 'react';
import styles from '@patternfly/react-styles/css/components/Page/page.mjs';
import { css } from '@patternfly/react-styles';
import BarsIcon from '@patternfly/react-icons/dist/esm/icons/bars-icon';
import { Button, ButtonVariant } from '../../../components/Button';
import { PageContextConsumer } from '../../../components/Page/PageContext';
export const PageHeader = (_a) => {
    var { className = '', logo = null, logoProps = null, logoComponent, headerTools = null, topNav = null, isNavOpen = true, role = undefined, showNavToggle = false, navToggleId = 'nav-toggle', onNavToggle = (_event) => undefined, 'aria-label': ariaLabel = 'Global navigation', 'aria-controls': ariaControls = null } = _a, props = __rest(_a, ["className", "logo", "logoProps", "logoComponent", "headerTools", "topNav", "isNavOpen", "role", "showNavToggle", "navToggleId", "onNavToggle", 'aria-label', 'aria-controls']);
    let LogoComponent = logoComponent;
    if (!logoComponent) {
        if ((logoProps === null || logoProps === void 0 ? void 0 : logoProps.href) !== undefined) {
            LogoComponent = 'a';
        }
        else {
            LogoComponent = 'span';
        }
    }
    return (React.createElement(PageContextConsumer, null, ({ isManagedSidebar, onSidebarToggle: managedOnSidebarToggle, isSidebarOpen: managedIsSidebarOpen }) => {
        const navToggle = isManagedSidebar ? managedOnSidebarToggle : onNavToggle;
        const navOpen = isManagedSidebar ? managedIsSidebarOpen : isNavOpen;
        return (React.createElement("header", Object.assign({ role: role, className: css(styles.pageHeader, className) }, props),
            (showNavToggle || logo) && (React.createElement("div", { className: css(styles.pageHeaderBrand) },
                showNavToggle && (React.createElement("div", { className: css(styles.pageHeaderBrandToggle) },
                    React.createElement(Button, { id: navToggleId, onClick: navToggle, "aria-label": ariaLabel, "aria-controls": ariaControls, "aria-expanded": navOpen ? 'true' : 'false', variant: ButtonVariant.plain },
                        React.createElement(BarsIcon, null)))),
                logo && (React.createElement(LogoComponent, Object.assign({ className: css(styles.pageHeaderBrandLink) }, logoProps), logo)))),
            topNav && React.createElement("div", { className: css(styles.pageHeaderNav) }, topNav),
            headerTools));
    }));
};
PageHeader.displayName = 'PageHeader';
//# sourceMappingURL=PageHeader.js.map