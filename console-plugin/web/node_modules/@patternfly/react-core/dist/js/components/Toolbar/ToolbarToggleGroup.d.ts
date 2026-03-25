import * as React from 'react';
import { ToolbarGroupProps } from './ToolbarGroup';
export interface ToolbarToggleGroupProps extends ToolbarGroupProps {
    /** Flag indicating when toggle group is expanded for non-managed toolbar toggle groups. */
    isExpanded?: boolean;
    /** Callback for toggle group click event for non-managed toolbar toggle groups. */
    onToggle?: (event: React.MouseEvent) => void;
    /** An icon to be rendered when the toggle group has collapsed down */
    toggleIcon: React.ReactNode;
    /** Controls when filters are shown and when the toggle button is hidden. */
    breakpoint: 'md' | 'lg' | 'xl' | '2xl';
    /** Visibility at various breakpoints. */
    visibility?: {
        default?: 'hidden' | 'visible';
        md?: 'hidden' | 'visible';
        lg?: 'hidden' | 'visible';
        xl?: 'hidden' | 'visible';
        '2xl'?: 'hidden' | 'visible';
    };
    /** Alignment at various breakpoints. */
    alignment?: {
        default?: 'alignRight' | 'alignLeft';
        md?: 'alignRight' | 'alignLeft';
        lg?: 'alignRight' | 'alignLeft';
        xl?: 'alignRight' | 'alignLeft';
        '2xl'?: 'alignRight' | 'alignLeft';
    };
    /** Spacers at various breakpoints. */
    spacer?: {
        default?: 'spacerNone' | 'spacerSm' | 'spacerMd' | 'spacerLg';
        md?: 'spacerNone' | 'spacerSm' | 'spacerMd' | 'spacerLg';
        lg?: 'spacerNone' | 'spacerSm' | 'spacerMd' | 'spacerLg';
        xl?: 'spacerNone' | 'spacerSm' | 'spacerMd' | 'spacerLg';
        '2xl'?: 'spacerNone' | 'spacerSm' | 'spacerMd' | 'spacerLg';
    };
    /** Space items at various breakpoints. */
    spaceItems?: {
        default?: 'spaceItemsNone' | 'spaceItemsSm' | 'spaceItemsMd' | 'spaceItemsLg';
        md?: 'spaceItemsNone' | 'spaceItemsSm' | 'spaceItemsMd' | 'spaceItemsLg';
        lg?: 'spaceItemsNone' | 'spaceItemsSm' | 'spaceItemsMd' | 'spaceItemsLg';
        xl?: 'spaceItemsNone' | 'spaceItemsSm' | 'spaceItemsMd' | 'spaceItemsLg';
        '2xl'?: 'spaceItemsNone' | 'spaceItemsSm' | 'spaceItemsMd' | 'spaceItemsLg';
    };
    /** Reference to a chip container group for filters inside the toolbar toggle group */
    chipContainerRef?: React.RefObject<any>;
    /** Optional callback for clearing all filters in the toolbar toggle group */
    clearAllFilters?: () => void;
    /** Flag indicating that the clear all filters button should be visible in the toolbar toggle group */
    showClearFiltersButton?: boolean;
    /** Text to display in the clear all filters button of the toolbar toggle group */
    clearFiltersButtonText?: string;
}
declare class ToolbarToggleGroup extends React.Component<ToolbarToggleGroupProps> {
    static displayName: string;
    toggleRef: React.RefObject<HTMLButtonElement>;
    expandableContentRef: React.RefObject<HTMLDivElement>;
    isContentPopup: () => boolean;
    render(): React.JSX.Element;
}
export { ToolbarToggleGroup };
//# sourceMappingURL=ToolbarToggleGroup.d.ts.map