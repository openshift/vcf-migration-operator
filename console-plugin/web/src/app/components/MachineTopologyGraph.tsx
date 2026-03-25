import * as React from 'react';
import {
  ColaLayout,
  ComponentFactory,
  DefaultEdge,
  DefaultGroup,
  DefaultNode,
  Graph,
  GraphComponent,
  Layout,
  LayoutFactory,
  Model,
  ModelKind,
  Node,
  NodeModel,
  NodeShape,
  NodeStatus,
  Visualization,
  VisualizationProvider,
  VisualizationSurface,
} from '@patternfly/react-topology';
import { ServerIcon, DesktopIcon } from '@patternfly/react-icons';
import type { MachineRow } from './MachineTopologyView';

/* ------------------------------------------------------------------ */
/* Constants                                                           */
/* ------------------------------------------------------------------ */

const NODE_DIAMETER = 75;

/** Badge colors by machine role. */
const BADGE_COLORS: Record<string, { bg: string; text: string; border: string }> = {
  master: { bg: '#F2F0FC', text: '#5752d1', border: '#CBC1FF' },
  worker: { bg: '#E7F1FA', text: '#06c', border: '#bee1f4' },
};

/* ------------------------------------------------------------------ */
/* Status mapping                                                      */
/* ------------------------------------------------------------------ */

/**
 * Derives a PatternFly NodeStatus from machine phase and node readiness.
 *
 * - Provisioning / Provisioned → info (blue spinner)
 * - Running + Node Ready        → success (green check)
 * - Running + Node Not Ready    → warning (yellow triangle)
 * - Failed / Deleting           → danger (red exclamation)
 * - Unknown / pending           → default (grey)
 */
const machineStatus = (phase: string, nodeReady: boolean | null): NodeStatus => {
  switch (phase) {
    case 'Provisioning':
    case 'Provisioned':
      return NodeStatus.info;
    case 'Running':
      if (nodeReady === true) return NodeStatus.success;
      if (nodeReady === false) return NodeStatus.warning;
      return NodeStatus.default;
    case 'Failed':
    case 'Deleting':
      return NodeStatus.danger;
    default:
      return NodeStatus.default;
  }
};

/* ------------------------------------------------------------------ */
/* Custom node                                                         */
/* ------------------------------------------------------------------ */

const CustomNode: React.FC<{ element: Node }> = ({ element }) => {
  const data = element.getData();
  const Icon = data.role === 'master' ? ServerIcon : DesktopIcon;
  const badge = BADGE_COLORS[data.role as string] ?? BADGE_COLORS.worker;

  return (
    <DefaultNode
      element={element}
      showStatusDecorator
      badge={data.role === 'master' ? 'CP' : 'W'}
      badgeColor={badge.bg}
      badgeTextColor={badge.text}
      badgeBorderColor={badge.border}
    >
      <g transform="translate(25, 25)">
        <Icon width={25} height={25} />
      </g>
    </DefaultNode>
  );
};

/* ------------------------------------------------------------------ */
/* Factories                                                           */
/* ------------------------------------------------------------------ */

const layoutFactory: LayoutFactory = (_type: string, graph: Graph): Layout | undefined =>
  new ColaLayout(graph, { layoutOnDrag: false });

const componentFactory: ComponentFactory = (kind: ModelKind, type: string) => {
  if (type === 'group') return DefaultGroup;
  switch (kind) {
    case ModelKind.graph:
      return GraphComponent;
    case ModelKind.node:
      return CustomNode as React.ComponentType<any>; // eslint-disable-line @typescript-eslint/no-explicit-any
    case ModelKind.edge:
      return DefaultEdge;
    default:
      return undefined;
  }
};

/* ------------------------------------------------------------------ */
/* Model builder                                                       */
/* ------------------------------------------------------------------ */

/**
 * Builds the PatternFly topology Model from machine rows.
 *
 * Groups are created per-vCenter. Each machine becomes a node inside
 * its vCenter group, color-coded by role (badge) and status-decorated
 * based on machine phase / node readiness.
 */
const buildTopologyModel = (rows: MachineRow[]): Model => {
  const groups = new Map<string, string[]>();
  const nodes: NodeModel[] = [];

  rows.forEach((row) => {
    const id = `machine-${row.machineName}`;
    const groupKey = row.vcenter || 'unknown';

    if (!groups.has(groupKey)) {
      groups.set(groupKey, []);
    }
    groups.get(groupKey)!.push(id);

    nodes.push({
      id,
      type: 'node',
      label: row.machineName,
      width: NODE_DIAMETER,
      height: NODE_DIAMETER,
      shape: row.role === 'master' ? NodeShape.hexagon : NodeShape.ellipse,
      status: machineStatus(row.machinePhase, row.nodeReady),
      data: {
        role: row.role,
        phase: row.machinePhase,
        nodeReady: row.nodeReady,
      },
    });
  });

  // Create group nodes for each vCenter.
  groups.forEach((children, vcenter) => {
    nodes.push({
      id: `group-${vcenter}`,
      children,
      type: 'group',
      group: true,
      label: vcenter,
      style: { padding: 40 },
    });
  });

  return {
    nodes,
    edges: [],
    graph: { id: 'machine-topology', type: 'graph', layout: 'Cola' },
  };
};

/* ------------------------------------------------------------------ */
/* Component                                                           */
/* ------------------------------------------------------------------ */

export interface MachineTopologyGraphProps {
  rows: MachineRow[];
}

export const MachineTopologyGraph: React.FC<MachineTopologyGraphProps> = ({ rows }) => {
  const controller = React.useMemo(() => {
    const viz = new Visualization();
    viz.registerLayoutFactory(layoutFactory);
    viz.registerComponentFactory(componentFactory);
    viz.fromModel(buildTopologyModel(rows), false);
    return viz;
  }, [rows]);

  return (
    <div style={{ height: 500, border: '1px solid var(--pf-v5-global--BorderColor--100)' }}>
      <VisualizationProvider controller={controller}>
        <VisualizationSurface />
      </VisualizationProvider>
    </div>
  );
};
