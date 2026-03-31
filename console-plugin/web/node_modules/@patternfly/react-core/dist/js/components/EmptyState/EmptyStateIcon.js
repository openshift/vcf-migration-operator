"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.EmptyStateIcon = void 0;
const tslib_1 = require("tslib");
const React = tslib_1.__importStar(require("react"));
const react_styles_1 = require("@patternfly/react-styles");
const empty_state_1 = tslib_1.__importDefault(require("@patternfly/react-styles/css/components/EmptyState/empty-state"));
const Spinner_1 = require("../Spinner");
const c_empty_state__icon_Color_1 = tslib_1.__importDefault(require('@patternfly/react-tokens/dist/js/c_empty_state__icon_Color'));
const isSpinner = (icon) => icon.type === Spinner_1.Spinner;
const EmptyStateIcon = (_a) => {
    var { className, icon: IconComponent, color } = _a, props = tslib_1.__rest(_a, ["className", "icon", "color"]);
    const iconIsSpinner = isSpinner(React.createElement(IconComponent, null));
    return (React.createElement("div", Object.assign({ className: (0, react_styles_1.css)(empty_state_1.default.emptyStateIcon) }, (color && !iconIsSpinner && { style: { [c_empty_state__icon_Color_1.default.name]: color } })),
        React.createElement(IconComponent, Object.assign({ className: className, "aria-hidden": !iconIsSpinner }, props))));
};
exports.EmptyStateIcon = EmptyStateIcon;
exports.EmptyStateIcon.displayName = 'EmptyStateIcon';
//# sourceMappingURL=EmptyStateIcon.js.map