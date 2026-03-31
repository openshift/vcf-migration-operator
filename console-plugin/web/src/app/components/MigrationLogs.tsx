import * as React from 'react';
import {
  Alert,
  Button,
  Bullseye,
  Spinner,
  Toolbar,
  ToolbarContent,
  ToolbarItem,
  Select,
  SelectOption,
  SelectList,
  MenuToggle,
} from '@patternfly/react-core';
import { consoleFetchJSON, consoleFetchText } from '@openshift-console/dynamic-plugin-sdk';

const OPERATOR_NAMESPACE = 'openshift-vcf-migration';
const TAIL_LINES = 500;
const REFRESH_INTERVAL_MS = 10000;

interface PodItem {
  metadata: { name: string };
}

export const MigrationLogs: React.FC = () => {
  const [pods, setPods] = React.useState<string[]>([]);
  const [selectedPod, setSelectedPod] = React.useState<string>('');
  const [podSelectOpen, setPodSelectOpen] = React.useState(false);
  const [logs, setLogs] = React.useState<string>('');
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const logRef = React.useRef<HTMLPreElement>(null);
  const intervalRef = React.useRef<ReturnType<typeof setInterval> | null>(null);

  React.useEffect(() => {
    let cancelled = false;
    const fetchPods = async () => {
      try {
        const data = (await consoleFetchJSON(
          `/api/kubernetes/api/v1/namespaces/${OPERATOR_NAMESPACE}/pods?labelSelector=control-plane%3Dcontroller-manager`,
        )) as { items: PodItem[] };
        const names = data.items?.map((p) => p.metadata.name) ?? [];
        if (!cancelled) {
          setPods(names);
          if (names.length > 0 && !selectedPod) setSelectedPod(names[0]);
        }
      } catch (e) {
        if (!cancelled) setError(`Failed to list operator pods: ${e instanceof Error ? e.message : String(e)}`);
      }
    };
    fetchPods();
    return () => { cancelled = true; };
  }, []);

  const fetchLogs = React.useCallback(async (podName: string) => {
    if (!podName) return;
    try {
      const raw = await consoleFetchText(
        `/api/kubernetes/api/v1/namespaces/${OPERATOR_NAMESPACE}/pods/${podName}/log?tailLines=${TAIL_LINES}&container=manager`,
      );
      setLogs(raw);
      setError(null);
    } catch (e) {
      setError(`Failed to fetch logs: ${e instanceof Error ? e.message : String(e)}`);
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    if (!selectedPod) return;
    setLoading(true);
    fetchLogs(selectedPod);
    intervalRef.current = setInterval(() => fetchLogs(selectedPod), REFRESH_INTERVAL_MS);
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [selectedPod, fetchLogs]);

  React.useEffect(() => {
    if (logRef.current) {
      logRef.current.scrollTop = logRef.current.scrollHeight;
    }
  }, [logs]);

  if (error && !logs) {
    return <Alert variant="warning" title="Logs unavailable" isInline>{error}</Alert>;
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <Toolbar>
        <ToolbarContent>
          <ToolbarItem>
            <Select
              isOpen={podSelectOpen}
              selected={selectedPod}
              onSelect={(_e, val) => { setSelectedPod(val as string); setPodSelectOpen(false); }}
              onOpenChange={setPodSelectOpen}
              toggle={(toggleRef) => (
                <MenuToggle ref={toggleRef} onClick={() => setPodSelectOpen((p) => !p)} isExpanded={podSelectOpen} style={{ minWidth: 300 }}>
                  {selectedPod || 'Select pod'}
                </MenuToggle>
              )}
            >
              <SelectList>
                {pods.map((p) => (
                  <SelectOption key={p} value={p}>{p}</SelectOption>
                ))}
              </SelectList>
            </Select>
          </ToolbarItem>
          <ToolbarItem>
            <Button variant="plain" onClick={() => { if (selectedPod) fetchLogs(selectedPod); }}>
              Refresh
            </Button>
          </ToolbarItem>
        </ToolbarContent>
      </Toolbar>
      {loading ? (
        <Bullseye><Spinner size="lg" aria-label="Loading logs" /></Bullseye>
      ) : (
        <pre
          ref={logRef}
          style={{
            flex: 1,
            overflow: 'auto',
            margin: 0,
            padding: '1rem',
            backgroundColor: 'var(--pf-v5-global--BackgroundColor--dark-300, #1b1d21)',
            color: 'var(--pf-v5-global--Color--light-100, #e0e0e0)',
            fontSize: '0.8125rem',
            lineHeight: 1.5,
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-all',
          }}
        >
          {logs || 'No logs available.'}
        </pre>
      )}
    </div>
  );
};
