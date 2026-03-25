import { useCallback, useState } from 'react';
import { consoleFetch } from '@openshift-console/dynamic-plugin-sdk';

const API_BASE = '/api/proxy/plugin/vcf-migration-console/vcf-migration-api';

export interface VSphereConnectParams {
  server: string;
  username?: string;
  password?: string;
  secretRef?: { name: string; namespace?: string };
  createSecret?: { name: string; namespace?: string };
}

export interface VSphereConnectResult {
  datacenters: string[];
  secretCreated?: { name: string; namespace: string };
  error?: string;
}

export function useVSphereConnect() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const connect = useCallback(async (params: VSphereConnectParams): Promise<VSphereConnectResult | null> => {
    setLoading(true);
    setError(null);
    try {
      const res = await consoleFetch(`${API_BASE}/vsphere/connect`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(params),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.error || res.statusText);
        return { datacenters: [], error: data.error || res.statusText };
      }
      return data as VSphereConnectResult;
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      setError(msg);
      return { datacenters: [], error: msg };
    } finally {
      setLoading(false);
    }
  }, []);

  return { connect, loading, error };
}

export function useVSphereList(
  endpoint: string,
  params: { server: string; datacenter: string; secretName?: string; secretNamespace?: string; username?: string; password?: string }
) {
  const [items, setItems] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchList = useCallback(async () => {
    if (!params.server || !params.datacenter) return;
    setLoading(true);
    setError(null);
    try {
      const searchParams = new URLSearchParams({
        server: params.server,
        datacenter: params.datacenter,
      });
      if (params.secretName) {
        searchParams.set('secretName', params.secretName);
        if (params.secretNamespace) searchParams.set('secretNamespace', params.secretNamespace);
      } else if (params.username && params.password) {
        searchParams.set('username', params.username);
        searchParams.set('password', params.password);
      }
      const res = await consoleFetch(`${API_BASE}/vsphere/${endpoint}?${searchParams}`);
      const data = await res.json();
      if (!res.ok) {
        setError(data.error || res.statusText);
        setItems([]);
        return;
      }
      setItems(data.items || []);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
      setItems([]);
    } finally {
      setLoading(false);
    }
  }, [endpoint, params.server, params.datacenter, params.secretName, params.secretNamespace, params.username, params.password]);

  return { items, loading, error, fetchList };
}
