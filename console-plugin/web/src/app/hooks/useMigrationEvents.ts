import { useEffect, useRef, useState, useCallback } from 'react';

export interface MigrationEvent {
  type: string;
  reason: string;
  message: string;
  count: number;
  lastTimestamp: string;
  involvedObject: { kind: string; name: string; namespace: string };
}

const MAX_EVENTS = 100;
const MAX_RETRIES = 5;
const BASE_RETRY_MS = 2000;

export function useMigrationEvents(namespace: string | null, name: string | null) {
  const [events, setEvents] = useState<MigrationEvent[]>([]);
  const [error, setError] = useState<string | null>(null);
  const eventSourceRef = useRef<EventSource | null>(null);
  const retriesRef = useRef(0);

  const connect = useCallback(() => {
    if (!namespace || !name) return;

    const url = `/api/proxy/plugin/vcf-migration-console/vcf-migration-api/events?namespace=${encodeURIComponent(namespace)}&name=${encodeURIComponent(name)}`;
    const es = new EventSource(url);
    eventSourceRef.current = es;

    es.onopen = () => {
      retriesRef.current = 0;
      setError(null);
    };

    es.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data) as MigrationEvent;
        setError(null);
        setEvents((prev) => {
          const next = [...prev];
          const idx = next.findIndex(
            (ev) =>
              ev.reason === data.reason &&
              ev.message === data.message &&
              ev.lastTimestamp === data.lastTimestamp
          );
          if (idx >= 0) {
            next[idx] = data;
          } else {
            next.unshift(data);
          }
          return next.slice(0, MAX_EVENTS);
        });
      } catch {
        // ignore parse errors
      }
    };

    es.onerror = () => {
      es.close();
      eventSourceRef.current = null;
      retriesRef.current += 1;
      if (retriesRef.current > MAX_RETRIES) {
        setError('Event stream connection lost. Reload the page to retry.');
        return;
      }
      const delay = BASE_RETRY_MS * Math.pow(2, retriesRef.current - 1);
      setTimeout(connect, delay);
    };
  }, [namespace, name]);

  useEffect(() => {
    setEvents([]);
    setError(null);
    retriesRef.current = 0;

    if (!namespace || !name) return;

    connect();

    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
    };
  }, [namespace, name, connect]);

  return { events, error };
}
