"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ApplicationLauncherText = void 0;
const tslib_1 = require("tslib");
const React = tslib_1.__importStar(require("react"));
const react_styles_1 = require("@patternfly/react-styles");
const app_launcher_1 = tslib_1.__importDefault(require("@patternfly/react-styles/css/components/AppLauncher/app-launcher"));
const ApplicationLauncherText = (_a) => {
    var { className = '', children } = _a, props = tslib_1.__rest(_a, ["className", "children"]);
    return (React.createElement("span", Object.assign({ className: (0, react_styles_1.css)(`${app_launcher_1.default.appLauncherMenuItem}-text`, className) }, props), children));
};
exports.ApplicationLauncherText = ApplicationLauncherText;
exports.ApplicationLauncherText.displayName = 'ApplicationLauncherText';
//# sourceMappingURL=ApplicationLauncherText.js.map