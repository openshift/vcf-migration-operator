import * as React from 'react';
import { EmptyStateIconProps } from './EmptyStateIcon';
export interface EmptyStateHeaderProps extends React.HTMLProps<HTMLDivElement> {
    /** Content rendered inside the empty state header, either in addition to or instead of the titleText prop */
    children?: React.ReactNode;
    /** Additional classes added to the empty state header */
    className?: string;
    /** Additional classes added to the title inside empty state header */
    titleClassName?: string;
    /** Text of the title inside empty state header, will be wrapped in headingLevel */
    titleText?: React.ReactNode;
    /** Empty state icon element to be rendered */
    icon?: React.ReactElement<EmptyStateIconProps>;
    /** The heading level to use, default is h1 */
    headingLevel?: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6';
}
export declare const EmptyStateHeader: React.FunctionComponent<EmptyStateHeaderProps>;
//# sourceMappingURL=EmptyStateHeader.d.ts.map