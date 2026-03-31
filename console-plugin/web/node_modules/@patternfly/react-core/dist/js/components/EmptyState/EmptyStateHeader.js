"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.EmptyStateHeader = void 0;
const tslib_1 = require("tslib");
const React = tslib_1.__importStar(require("react"));
const react_styles_1 = require("@patternfly/react-styles");
const empty_state_1 = tslib_1.__importDefault(require("@patternfly/react-styles/css/components/EmptyState/empty-state"));
const EmptyStateHeader = (_a) => {
    var { children, className, titleClassName, titleText, headingLevel: HeadingLevel = 'h1', icon } = _a, props = tslib_1.__rest(_a, ["children", "className", "titleClassName", "titleText", "headingLevel", "icon"]);
    return (React.createElement("div", Object.assign({ className: (0, react_styles_1.css)(`${empty_state_1.default.emptyState}__header`, className) }, props),
        icon,
        (titleText || children) && (React.createElement("div", { className: (0, react_styles_1.css)(`${empty_state_1.default.emptyState}__title`) },
            titleText && (React.createElement(HeadingLevel, { className: (0, react_styles_1.css)(empty_state_1.default.emptyStateTitleText, titleClassName) }, titleText)),
            children))));
};
exports.EmptyStateHeader = EmptyStateHeader;
exports.EmptyStateHeader.displayName = 'EmptyStateHeader';
//# sourceMappingURL=EmptyStateHeader.js.map