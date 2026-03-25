import * as React from 'react';
import {
  EmptyState,
  EmptyStateHeader,
  EmptyStateIcon,
  EmptyStateBody,
  Label,
  Alert,
} from '@patternfly/react-core';
import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
} from '@patternfly/react-table';
import { InfoCircleIcon } from '@patternfly/react-icons';
import { useMigrationEvents, MigrationEvent } from '../hooks/useMigrationEvents';

export interface EventStreamProps {
  namespace: string;
  name: string;
}

const formatTimestamp = (ts: string): string => {
  if (!ts) return '-';
  const d = new Date(ts);
  return d.toLocaleString();
};

const getEventColor = (type: string): 'blue' | 'orange' | 'grey' => {
  switch (type) {
    case 'Normal':
      return 'blue';
    case 'Warning':
      return 'orange';
    default:
      return 'grey';
  }
};

export const EventStream: React.FC<EventStreamProps> = (props) => {
  const { events, error } = useMigrationEvents(props.namespace, props.name);

  if (error) {
    return (
      <Alert variant="warning" title="Event stream unavailable" isInline>
        {error}
      </Alert>
    );
  }

  if (events.length === 0) {
    return (
      <EmptyState>
        <EmptyStateHeader
          titleText="No events"
          headingLevel="h4"
          icon={<EmptyStateIcon icon={InfoCircleIcon} />}
        />
        <EmptyStateBody>
          No events have been recorded for this migration yet.
        </EmptyStateBody>
      </EmptyState>
    );
  }

  return (
    <Table aria-label="Migration events" variant="compact">
      <Thead>
        <Tr>
          <Th>Type</Th>
          <Th>Reason</Th>
          <Th>Message</Th>
          <Th>Last seen</Th>
          <Th>Count</Th>
        </Tr>
      </Thead>
      <Tbody>
        {events.map((ev: MigrationEvent, i: number) => (
          <Tr key={`${ev.lastTimestamp}-${ev.reason}-${i}`}>
            <Td dataLabel="Type">
              <Label color={getEventColor(ev.type)} isCompact>{ev.type}</Label>
            </Td>
            <Td dataLabel="Reason">{ev.reason}</Td>
            <Td dataLabel="Message">{ev.message}</Td>
            <Td dataLabel="Last seen">{formatTimestamp(ev.lastTimestamp)}</Td>
            <Td dataLabel="Count">{ev.count}</Td>
          </Tr>
        ))}
      </Tbody>
    </Table>
  );
};
