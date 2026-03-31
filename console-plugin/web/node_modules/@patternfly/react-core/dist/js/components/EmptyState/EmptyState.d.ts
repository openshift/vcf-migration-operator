import * as React from 'react';
export declare enum EmptyStateVariant {
    'xs' = "xs",
    sm = "sm",
    lg = "lg",
    'xl' = "xl",
    full = "full"
}
export interface EmptyStateProps extends React.HTMLProps<HTMLDivElement> {
    /** Additional classes added to the empty state */
    className?: string;
    /** Content rendered inside the empty state */
    children: React.ReactNode;
    /** Modifies empty state max-width and sizes of icon, title and body */
    variant?: 'xs' | 'sm' | 'lg' | 'xl' | 'full';
    /** Cause component to consume the available height of its container */
    isFullHeight?: boolean;
}
export declare const EmptyState: React.FunctionComponent<EmptyStateProps>;
//# sourceMappingURL=EmptyState.d.ts.map