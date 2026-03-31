import * as React from 'react';
import {
  MenuToggle,
  Select,
  SelectList,
  SelectOption,
  TextInputGroup,
  TextInputGroupMain,
  TextInputGroupUtilities,
  Button,
  Spinner,
} from '@patternfly/react-core';
import { TimesIcon } from '@patternfly/react-icons';

export interface TypeaheadSelectProps {
  id: string;
  items: string[];
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  isDisabled?: boolean;
  isLoading?: boolean;
  'aria-label'?: string;
}

export const TypeaheadSelect: React.FC<TypeaheadSelectProps> = ({
  id,
  items,
  value,
  onChange,
  placeholder = 'Select...',
  isDisabled = false,
  isLoading = false,
  ...props
}) => {
  const [isOpen, setIsOpen] = React.useState(false);
  const [filterValue, setFilterValue] = React.useState('');
  const textInputRef = React.useRef<HTMLInputElement>(null);

  const filtered = React.useMemo(() => {
    if (!filterValue) return items;
    const lower = filterValue.toLowerCase();
    return items.filter((item) => item.toLowerCase().includes(lower));
  }, [items, filterValue]);

  const handleSelect = React.useCallback(
    (_e: React.MouseEvent | undefined, val: string | number | undefined) => {
      onChange(val as string);
      setIsOpen(false);
      setFilterValue('');
    },
    [onChange],
  );

  const handleClear = React.useCallback(() => {
    onChange('');
    setFilterValue('');
    textInputRef.current?.focus();
  }, [onChange]);

  const handleToggle = React.useCallback(() => {
    if (!isDisabled) setIsOpen((prev) => !prev);
  }, [isDisabled]);

  const handleInputChange = React.useCallback(
    (_e: React.FormEvent, val: string) => {
      setFilterValue(val);
      if (!isOpen) setIsOpen(true);
    },
    [isOpen],
  );

  const displayValue = isOpen ? filterValue : value;

  const toggle = React.useCallback(
    (toggleRef: React.Ref<MenuToggleElement>) => (
      <MenuToggle
        ref={toggleRef}
        variant="typeahead"
        onClick={handleToggle}
        isExpanded={isOpen}
        isDisabled={isDisabled}
        isFullWidth
      >
        <TextInputGroup isPlain>
          <TextInputGroupMain
            value={displayValue}
            onClick={handleToggle}
            onChange={handleInputChange}
            autoComplete="off"
            innerRef={textInputRef}
            placeholder={placeholder}
            aria-label={props['aria-label'] ?? placeholder}
            id={id}
          />
          {(value || filterValue) && !isDisabled && (
            <TextInputGroupUtilities>
              {isLoading && <Spinner size="sm" />}
              <Button variant="plain" onClick={handleClear} aria-label="Clear">
                <TimesIcon />
              </Button>
            </TextInputGroupUtilities>
          )}
        </TextInputGroup>
      </MenuToggle>
    ),
    [handleToggle, isOpen, isDisabled, displayValue, handleInputChange, placeholder, props, id, value, filterValue, isLoading, handleClear],
  );

  return (
    <Select
      id={`${id}-select`}
      isOpen={isOpen}
      selected={value}
      onSelect={handleSelect}
      onOpenChange={setIsOpen}
      toggle={toggle}
      variant="typeahead"
      isScrollable
      maxMenuHeight="300px"
    >
      <SelectList>
        {isLoading ? (
          <SelectOption isDisabled value="loading">Loading...</SelectOption>
        ) : filtered.length === 0 ? (
          <SelectOption isDisabled value="no-results">No results found</SelectOption>
        ) : (
          filtered.map((item) => (
            <SelectOption key={item} value={item}>
              {item}
            </SelectOption>
          ))
        )}
      </SelectList>
    </Select>
  );
};

type MenuToggleElement = HTMLButtonElement;
