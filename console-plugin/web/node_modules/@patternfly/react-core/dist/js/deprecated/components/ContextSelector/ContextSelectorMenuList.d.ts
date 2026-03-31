import * as React from 'react';
export interface ContextSelectorMenuListProps {
    /** Content rendered inside the context selector menu */
    children?: React.ReactNode;
    /** Classess applied to root element of context selector menu */
    className?: string;
    /** Flag to indicate if context selector menu is opened */
    isOpen?: boolean;
    /** Adds an accessible label to the context selector menu. */
    'aria-label'?: string;
}
declare class ContextSelectorMenuList extends React.Component<ContextSelectorMenuListProps> {
    static displayName: string;
    static defaultProps: ContextSelectorMenuListProps;
    refsCollection: any;
    sendRef: (index: number, ref: any) => void;
    extendChildren(): React.ReactElement<any, string | React.JSXElementConstructor<any>>[];
    render: () => React.JSX.Element;
}
export { ContextSelectorMenuList };
//# sourceMappingURL=ContextSelectorMenuList.d.ts.map