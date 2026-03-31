import { __rest } from "tslib";
import * as React from 'react';
import { css } from '@patternfly/react-styles';
import { Button } from '../Button';
import { Tooltip } from '../Tooltip';
import TimesIcon from '@patternfly/react-icons/dist/esm/icons/times-icon';
import styles from '@patternfly/react-styles/css/components/Chip/chip.mjs';
import { GenerateId } from '../../helpers/GenerateId/GenerateId';
import { getOUIAProps, getDefaultOUIAId } from '../../helpers';
import cssChipTextMaxWidth from '@patternfly/react-tokens/dist/esm/c_chip__text_MaxWidth';
class Chip extends React.Component {
    constructor(props) {
        super(props);
        this.span = React.createRef();
        this.setChipStyle = () => ({
            [cssChipTextMaxWidth.name]: this.props.textMaxWidth
        });
        this.renderOverflowChip = () => {
            const _a = this.props, { badge, children, className, onClick, ouiaId, textMaxWidth, style, component, 
            // eslint-disable-next-line @typescript-eslint/no-unused-vars
            tooltipPosition, 
            // eslint-disable-next-line @typescript-eslint/no-unused-vars
            isOverflowChip, 
            // eslint-disable-next-line @typescript-eslint/no-unused-vars
            closeBtnAriaLabel, 
            // eslint-disable-next-line @typescript-eslint/no-unused-vars
            isReadOnly } = _a, props = __rest(_a, ["badge", "children", "className", "onClick", "ouiaId", "textMaxWidth", "style", "component", "tooltipPosition", "isOverflowChip", "closeBtnAriaLabel", "isReadOnly"]);
            const Component = component;
            return (React.createElement(Component, Object.assign({ onClick: onClick }, (textMaxWidth && Object.assign({ style: this.setChipStyle() }, style)), { className: css(styles.chip, styles.modifiers.overflow, className) }, (component === 'button' ? { type: 'button' } : {}), getOUIAProps('OverflowChip', ouiaId !== undefined ? ouiaId : this.state.ouiaStateId), props),
                React.createElement("span", { className: css(styles.chipContent) },
                    React.createElement("span", { className: css(styles.chipText) }, children),
                    badge && badge)));
        };
        this.renderChip = (randomId) => {
            const { children, tooltipPosition } = this.props;
            if (this.state.isTooltipVisible) {
                return (React.createElement(Tooltip, { position: tooltipPosition, content: children }, this.renderInnerChip(randomId)));
            }
            return this.renderInnerChip(randomId);
        };
        this.state = {
            isTooltipVisible: false,
            ouiaStateId: getDefaultOUIAId(Chip.displayName)
        };
    }
    componentDidMount() {
        this.setState({
            isTooltipVisible: Boolean(this.span.current && this.span.current.offsetWidth < this.span.current.scrollWidth)
        });
    }
    componentDidUpdate(_prevProps, prevState) {
        const nextIsTooltipVisible = Boolean(this.span.current && this.span.current.offsetWidth < this.span.current.scrollWidth);
        if (prevState.isTooltipVisible !== nextIsTooltipVisible) {
            this.setState({
                isTooltipVisible: nextIsTooltipVisible
            });
        }
    }
    renderInnerChip(id) {
        const _a = this.props, { badge, children, className, onClick, closeBtnAriaLabel, isReadOnly, 
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        isOverflowChip, 
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        tooltipPosition, 
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        id: idProp, component, ouiaId, textMaxWidth } = _a, props = __rest(_a, ["badge", "children", "className", "onClick", "closeBtnAriaLabel", "isReadOnly", "isOverflowChip", "tooltipPosition", "id", "component", "ouiaId", "textMaxWidth"]);
        const Component = component;
        return (React.createElement(Component, Object.assign({}, (textMaxWidth && {
            style: this.setChipStyle()
        }), { className: css(styles.chip, className) }, (this.state.isTooltipVisible && { tabIndex: 0 }), getOUIAProps(Chip.displayName, ouiaId !== undefined ? ouiaId : this.state.ouiaStateId), props),
            React.createElement("span", { className: css(styles.chipContent) },
                React.createElement("span", { ref: this.span, className: css(styles.chipText), id: id }, children),
                badge && badge),
            !isReadOnly && (React.createElement("span", { className: css(styles.chipActions) },
                React.createElement(Button, { onClick: onClick, variant: "plain", "aria-label": closeBtnAriaLabel, id: `remove_${id}`, "aria-labelledby": `remove_${id} ${id}`, ouiaId: ouiaId || closeBtnAriaLabel },
                    React.createElement(TimesIcon, { "aria-hidden": "true" }))))));
    }
    render() {
        const { isOverflowChip } = this.props;
        return (React.createElement(GenerateId, null, (randomId) => (isOverflowChip ? this.renderOverflowChip() : this.renderChip(this.props.id || randomId))));
    }
}
Chip.displayName = 'Chip';
Chip.defaultProps = {
    closeBtnAriaLabel: 'close',
    className: '',
    isOverflowChip: false,
    isReadOnly: false,
    tooltipPosition: 'top',
    onClick: (_e) => undefined,
    component: 'div'
};
export { Chip };
//# sourceMappingURL=Chip.js.map