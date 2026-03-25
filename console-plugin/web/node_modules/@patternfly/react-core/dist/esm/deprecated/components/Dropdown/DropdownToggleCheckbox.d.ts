import * as React from 'react';
import { PickOptional } from '../../../helpers/typeUtils';
import { OUIAProps } from '../../../helpers';
export interface DropdownToggleCheckboxProps extends Omit<React.HTMLProps<HTMLInputElement>, 'type' | 'onChange' | 'disabled' | 'checked'>, OUIAProps {
    /** Additional classes added to the DropdownToggleCheckbox */
    className?: string;
    /** Flag to show if the checkbox selection is valid or invalid */
    isValid?: boolean;
    /** Flag to show if the checkbox is disabled */
    isDisabled?: boolean;
    /** Flag to show if the checkbox is checked */
    isChecked?: boolean | null;
    /** Flag to show if the checkbox is in progress */
    isInProgress?: boolean | null;
    /** Alternate Flag to show if the checkbox is checked */
    checked?: boolean | null;
    /** A callback for when the checkbox selection changes */
    onChange?: (event: React.FormEvent<HTMLInputElement>, checked: boolean) => void;
    /** Element to be rendered inside the <span> */
    children?: React.ReactNode;
    /** Id of the checkbox */
    id: string;
    /** Aria-label of the checkbox */
    'aria-label': string;
    /** Text describing current loading status or progress */
    defaultProgressAriaValueText?: string;
    /** Aria-label for the default progress icon to describe what is loading */
    defaultProgressAriaLabel?: string;
    /** Value to overwrite the randomly generated data-ouia-component-id.*/
    ouiaId?: number | string;
    /** Set the value of data-ouia-safe. Only set to true when the component is in a static state, i.e. no animations are occurring. At all other times, this value must be false. */
    ouiaSafe?: boolean;
}
declare class DropdownToggleCheckbox extends React.Component<DropdownToggleCheckboxProps, {
    ouiaStateId: string;
}> {
    static displayName: string;
    static defaultProps: PickOptional<DropdownToggleCheckboxProps>;
    constructor(props: DropdownToggleCheckboxProps);
    handleChange: (event: React.FormEvent<HTMLInputElement>) => void;
    calculateChecked: () => boolean;
    render(): React.JSX.Element;
}
export { DropdownToggleCheckbox };
//# sourceMappingURL=DropdownToggleCheckbox.d.ts.map