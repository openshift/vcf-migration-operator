(self["webpackChunkvcf_migration_console"] = self["webpackChunkvcf_migration_console"] || []).push([["exposed-migrationPlugin"],{

/***/ 4475
(__unused_webpack_module, __webpack_exports__, __webpack_require__) {

"use strict";
// ESM COMPAT FLAG
__webpack_require__.r(__webpack_exports__);

// EXPORTS
__webpack_require__.d(__webpack_exports__, {
  MigrationDetailPage: () => (/* reexport */ MigrationDetailPage),
  MigrationListPage: () => (/* reexport */ MigrationListPage),
  MigrationWizard: () => (/* reexport */ MigrationWizard)
});

// EXTERNAL MODULE: ./node_modules/react/jsx-runtime.js
var jsx_runtime = __webpack_require__(4848);
// EXTERNAL MODULE: consume shared module (default) react@^17.0.1 (singleton)
var consume_shared_module_default_react_17_0_singleton_ = __webpack_require__(8893);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Page@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Page/index.js)
var index_js_ = __webpack_require__(2984);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Title@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Title/index.js)
var Title_index_js_ = __webpack_require__(3068);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Button@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Button/index.js)
var Button_index_js_ = __webpack_require__(2982);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Toolbar@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Toolbar/index.js)
var Toolbar_index_js_ = __webpack_require__(1176);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/EmptyState@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/EmptyState/index.js)
var EmptyState_index_js_ = __webpack_require__(5010);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Spinner@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Spinner/index.js)
var Spinner_index_js_ = __webpack_require__(9704);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/layouts/Bullseye@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/layouts/Bullseye/index.js)
var Bullseye_index_js_ = __webpack_require__(5464);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Label@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Label/index.js)
var Label_index_js_ = __webpack_require__(3592);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Alert@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Alert/index.js)
var Alert_index_js_ = __webpack_require__(3780);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Dropdown@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Dropdown/index.js)
var Dropdown_index_js_ = __webpack_require__(7152);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Divider@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Divider/index.js)
var Divider_index_js_ = __webpack_require__(2832);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/MenuToggle@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/MenuToggle/index.js)
var MenuToggle_index_js_ = __webpack_require__(3832);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-table/dist/dynamic/components/Table@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-table/dist/esm/components/Table/index.js)
var Table_index_js_ = __webpack_require__(8272);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-icons/dist/dynamic/icons/cubes-icon@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-icons/dist/esm/icons/cubes-icon.js)
var cubes_icon_js_ = __webpack_require__(3831);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-icons/dist/dynamic/icons/ellipsis-v-icon@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-icons/dist/esm/icons/ellipsis-v-icon.js)
var ellipsis_v_icon_js_ = __webpack_require__(7567);
// EXTERNAL MODULE: consume shared module (default) react-router-dom@~5.3 (singleton)
var consume_shared_module_default_react_router_dom_5_singleton_ = __webpack_require__(9359);
// EXTERNAL MODULE: consume shared module (default) @openshift-console/dynamic-plugin-sdk@^1.8.0 (singleton)
var dynamic_plugin_sdk_1_8_singleton_ = __webpack_require__(2385);
;// ./src/models.ts
const VmwareCloudFoundationMigrationModel = {
    kind: 'VmwareCloudFoundationMigration',
    label: 'VmwareCloudFoundationMigration',
    labelPlural: 'VmwareCloudFoundationMigrations',
    apiGroup: 'migration.openshift.io',
    apiVersion: 'v1alpha1',
    plural: 'vmwarecloudfoundationmigrations',
    abbr: 'vcfm',
    namespaced: true,
    crd: true,
};
const MachineSetModel = {
    kind: 'MachineSet',
    label: 'MachineSet',
    labelPlural: 'MachineSets',
    apiGroup: 'machine.openshift.io',
    apiVersion: 'v1beta1',
    plural: 'machinesets',
    abbr: 'ms',
    namespaced: true,
};
const MachineModel = {
    kind: 'Machine',
    label: 'Machine',
    labelPlural: 'Machines',
    apiGroup: 'machine.openshift.io',
    apiVersion: 'v1beta1',
    plural: 'machines',
    abbr: 'm',
    namespaced: true,
};
const ControlPlaneMachineSetModel = {
    kind: 'ControlPlaneMachineSet',
    label: 'ControlPlaneMachineSet',
    labelPlural: 'ControlPlaneMachineSets',
    apiGroup: 'machine.openshift.io',
    apiVersion: 'v1',
    plural: 'controlplanemachinesets',
    abbr: 'cpms',
    namespaced: true,
};
const NodeModel = {
    kind: 'Node',
    label: 'Node',
    labelPlural: 'Nodes',
    apiGroup: '',
    apiVersion: 'v1',
    plural: 'nodes',
    abbr: 'n',
    namespaced: false,
};

;// ./src/app/pages/MigrationListPage.tsx


































const migrationGVK = {
    group: 'migration.openshift.io',
    version: 'v1alpha1',
    kind: 'VmwareCloudFoundationMigration',
};
const migrationStates = ['Pending', 'Running', 'Paused'];
const getStateColor = (state) => {
    switch (state) {
        case 'Running':
            return 'blue';
        case 'Paused':
            return 'orange';
        default:
            return 'grey';
    }
};
const getReadyColor = (status) => {
    switch (status) {
        case 'True':
            return 'green';
        case 'False':
            return 'red';
        default:
            return 'grey';
    }
};
const formatAge = (timestamp) => {
    if (!timestamp || typeof timestamp !== 'string')
        return '-';
    const created = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - created.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    if (diffMins < 60)
        return `${diffMins}m`;
    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24)
        return `${diffHours}h`;
    const diffDays = Math.floor(diffHours / 24);
    return `${diffDays}d`;
};
const RowActions = ({ migration, onError }) => {
    const [isOpen, setIsOpen] = consume_shared_module_default_react_17_0_singleton_.useState(false);
    const handleSetState = consume_shared_module_default_react_17_0_singleton_.useCallback(async (state) => {
        setIsOpen(false);
        try {
            await (0,dynamic_plugin_sdk_1_8_singleton_.k8sPatch)({
                model: VmwareCloudFoundationMigrationModel,
                resource: migration,
                data: [{ op: 'replace', path: '/spec/state', value: state }],
            });
        }
        catch (e) {
            onError(`Failed to set state to ${state}: ${e instanceof Error ? e.message : String(e)}`);
        }
    }, [migration, onError]);
    const handleDelete = consume_shared_module_default_react_17_0_singleton_.useCallback(async () => {
        setIsOpen(false);
        try {
            await (0,dynamic_plugin_sdk_1_8_singleton_.k8sDelete)({
                model: VmwareCloudFoundationMigrationModel,
                resource: migration,
            });
        }
        catch (e) {
            onError(`Failed to delete: ${e instanceof Error ? e.message : String(e)}`);
        }
    }, [migration, onError]);
    return ((0,jsx_runtime.jsx)(Dropdown_index_js_.Dropdown, { isOpen: isOpen, onSelect: () => setIsOpen(false), onOpenChange: setIsOpen, toggle: (toggleRef) => ((0,jsx_runtime.jsx)(MenuToggle_index_js_.MenuToggle, { ref: toggleRef, variant: "plain", onClick: (e) => {
                e.stopPropagation();
                setIsOpen((prev) => !prev);
            }, isExpanded: isOpen, "aria-label": "Actions", children: (0,jsx_runtime.jsx)(ellipsis_v_icon_js_.EllipsisVIcon, {}) })), popperProps: { position: 'right' }, children: (0,jsx_runtime.jsxs)(Dropdown_index_js_.DropdownList, { children: [migrationStates.map((state) => ((0,jsx_runtime.jsxs)(Dropdown_index_js_.DropdownItem, { onClick: (e) => {
                        e.stopPropagation();
                        handleSetState(state);
                    }, isDisabled: migration.spec.state === state, description: migration.spec.state === state ? 'Current state' : undefined, children: ["Set ", state] }, state))), (0,jsx_runtime.jsx)(Divider_index_js_.Divider, {}), (0,jsx_runtime.jsx)(Dropdown_index_js_.DropdownItem, { onClick: (e) => {
                        e.stopPropagation();
                        handleDelete();
                    }, isDanger: true, children: "Delete" }, "delete")] }) }));
};
const MigrationListPage = () => {
    const history = (0,consume_shared_module_default_react_router_dom_5_singleton_.useHistory)();
    const [actionError, setActionError] = consume_shared_module_default_react_17_0_singleton_.useState(null);
    const [migrations, loaded, loadError] = (0,dynamic_plugin_sdk_1_8_singleton_.useK8sWatchResource)({
        groupVersionKind: migrationGVK,
        isList: true,
        namespaced: true,
    });
    const getReadyCondition = (m) => {
        const cond = m.status?.conditions?.find((c) => c.type === 'Ready');
        return cond?.status ?? 'Unknown';
    };
    return ((0,jsx_runtime.jsxs)(jsx_runtime.Fragment, { children: [(0,jsx_runtime.jsx)(index_js_.PageSection, { variant: "light", children: (0,jsx_runtime.jsx)(Toolbar_index_js_.Toolbar, { children: (0,jsx_runtime.jsxs)(Toolbar_index_js_.ToolbarContent, { children: [(0,jsx_runtime.jsx)(Toolbar_index_js_.ToolbarItem, { children: (0,jsx_runtime.jsx)(Title_index_js_.Title, { headingLevel: "h1", children: "VCF Migrations" }) }), (0,jsx_runtime.jsx)(Toolbar_index_js_.ToolbarItem, { align: { default: 'alignRight' }, children: (0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "primary", onClick: () => history.push('/vcf-migration/create'), children: "Create migration" }) })] }) }) }), (0,jsx_runtime.jsxs)(index_js_.PageSection, { children: [actionError && ((0,jsx_runtime.jsx)(Alert_index_js_.Alert, { variant: "danger", title: "Action failed", isInline: true, className: "pf-v5-u-mb-md", actionClose: (0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "plain", onClick: () => setActionError(null), children: "Dismiss" }), children: actionError })), loadError && ((0,jsx_runtime.jsx)(Alert_index_js_.Alert, { variant: "danger", title: "Failed to load migrations", isInline: true, className: "pf-v5-u-mb-md", children: String(loadError) })), !loaded && !loadError && ((0,jsx_runtime.jsx)(Bullseye_index_js_.Bullseye, { children: (0,jsx_runtime.jsx)(Spinner_index_js_.Spinner, { size: "xl", "aria-label": "Loading migrations" }) })), loaded && !loadError && (!migrations || migrations.length === 0) && ((0,jsx_runtime.jsxs)(EmptyState_index_js_.EmptyState, { children: [(0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateHeader, { titleText: "No migrations", headingLevel: "h4", icon: (0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateIcon, { icon: cubes_icon_js_.CubesIcon }) }), (0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateBody, { children: "No VCF migrations have been created yet. Create a migration to begin moving your OpenShift cluster to a new vCenter." }), (0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateFooter, { children: (0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateActions, { children: (0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "primary", onClick: () => history.push('/vcf-migration/create'), children: "Create migration" }) }) })] })), loaded && !loadError && migrations?.length > 0 && ((0,jsx_runtime.jsxs)(Table_index_js_.Table, { "aria-label": "Migrations table", children: [(0,jsx_runtime.jsx)(Table_index_js_.Thead, { children: (0,jsx_runtime.jsxs)(Table_index_js_.Tr, { children: [(0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "Name" }), (0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "Namespace" }), (0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "State" }), (0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "Ready" }), (0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "Age" }), (0,jsx_runtime.jsx)(Table_index_js_.Th, { screenReaderText: "Actions" })] }) }), (0,jsx_runtime.jsx)(Table_index_js_.Tbody, { children: migrations.map((m) => {
                                    const readyStatus = getReadyCondition(m);
                                    return ((0,jsx_runtime.jsxs)(Table_index_js_.Tr, { isClickable: true, onRowClick: () => history.push(`/vcf-migration/ns/${m.metadata.namespace}/${m.metadata.name}`), children: [(0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "Name", children: m.metadata.name }), (0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "Namespace", children: m.metadata.namespace }), (0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "State", children: (0,jsx_runtime.jsx)(Label_index_js_.Label, { color: getStateColor(m.spec.state), children: m.spec.state }) }), (0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "Ready", children: (0,jsx_runtime.jsx)(Label_index_js_.Label, { color: getReadyColor(readyStatus), children: readyStatus }) }), (0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "Age", children: formatAge(m.metadata.creationTimestamp) }), (0,jsx_runtime.jsx)(Table_index_js_.Td, { isActionCell: true, children: (0,jsx_runtime.jsx)(RowActions, { migration: m, onError: setActionError }) })] }, `${m.metadata.namespace}-${m.metadata.name}`));
                                }) })] }))] })] }));
};

// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Wizard@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Wizard/index.js)
var Wizard_index_js_ = __webpack_require__(6544);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Form@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Form/index.js)
var Form_index_js_ = __webpack_require__(7178);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/HelperText@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/HelperText/index.js)
var HelperText_index_js_ = __webpack_require__(8152);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/TextInput@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/TextInput/index.js)
var TextInput_index_js_ = __webpack_require__(3168);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Checkbox@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Checkbox/index.js)
var Checkbox_index_js_ = __webpack_require__(8432);
;// ./src/app/components/wizard/CredentialsStep.tsx










const CredentialsStep = (props) => ((0,jsx_runtime.jsxs)(Form_index_js_.Form, { children: [(0,jsx_runtime.jsxs)(Form_index_js_.FormSection, { title: "Migration details", children: [(0,jsx_runtime.jsxs)(Form_index_js_.FormGroup, { label: "Migration name", isRequired: true, fieldId: "migration-name", children: [(0,jsx_runtime.jsx)(TextInput_index_js_.TextInput, { id: "migration-name", value: props.migrationName, onChange: (_e, v) => props.onMigrationNameChange(v), placeholder: "vcf-migration" }), (0,jsx_runtime.jsx)(Form_index_js_.FormHelperText, { children: (0,jsx_runtime.jsx)(HelperText_index_js_.HelperText, { children: (0,jsx_runtime.jsx)(HelperText_index_js_.HelperTextItem, { children: "A unique name for this migration resource" }) }) })] }), (0,jsx_runtime.jsxs)(Form_index_js_.FormGroup, { label: "Namespace", isRequired: true, fieldId: "migration-namespace", children: [(0,jsx_runtime.jsx)(TextInput_index_js_.TextInput, { id: "migration-namespace", value: props.migrationNamespace, onChange: (_e, v) => props.onMigrationNamespaceChange(v) }), (0,jsx_runtime.jsx)(Form_index_js_.FormHelperText, { children: (0,jsx_runtime.jsx)(HelperText_index_js_.HelperText, { children: (0,jsx_runtime.jsx)(HelperText_index_js_.HelperTextItem, { children: "The namespace where the migration resource will be created" }) }) })] })] }), (0,jsx_runtime.jsxs)(Form_index_js_.FormSection, { title: "Target vCenter", children: [(0,jsx_runtime.jsxs)(Form_index_js_.FormGroup, { label: "vCenter server", isRequired: true, fieldId: "server", children: [(0,jsx_runtime.jsx)(TextInput_index_js_.TextInput, { id: "server", value: props.server, onChange: (_e, v) => props.onServerChange(v), placeholder: "vcenter.example.com" }), (0,jsx_runtime.jsx)(Form_index_js_.FormHelperText, { children: (0,jsx_runtime.jsx)(HelperText_index_js_.HelperText, { children: (0,jsx_runtime.jsx)(HelperText_index_js_.HelperTextItem, { children: "FQDN or IP address of the target vCenter server" }) }) })] }), (0,jsx_runtime.jsx)(Form_index_js_.FormGroup, { fieldId: "use-secret", children: (0,jsx_runtime.jsx)(Checkbox_index_js_.Checkbox, { id: "use-secret", label: "Use existing secret for credentials", isChecked: props.useSecretRef, onChange: (_e, v) => props.onUseSecretRefChange(v), description: "Reference an existing Kubernetes secret containing vCenter credentials" }) }), props.useSecretRef ? ((0,jsx_runtime.jsxs)(jsx_runtime.Fragment, { children: [(0,jsx_runtime.jsx)(Form_index_js_.FormGroup, { label: "Secret name", isRequired: true, fieldId: "secret-name", children: (0,jsx_runtime.jsx)(TextInput_index_js_.TextInput, { id: "secret-name", value: props.secretName, onChange: (_e, v) => props.onSecretNameChange(v) }) }), (0,jsx_runtime.jsx)(Form_index_js_.FormGroup, { label: "Secret namespace", fieldId: "secret-namespace", children: (0,jsx_runtime.jsx)(TextInput_index_js_.TextInput, { id: "secret-namespace", value: props.secretNamespace, onChange: (_e, v) => props.onSecretNamespaceChange(v) }) })] })) : ((0,jsx_runtime.jsxs)(jsx_runtime.Fragment, { children: [(0,jsx_runtime.jsx)(Form_index_js_.FormGroup, { label: "Username", isRequired: true, fieldId: "username", children: (0,jsx_runtime.jsx)(TextInput_index_js_.TextInput, { id: "username", value: props.username, onChange: (_e, v) => props.onUsernameChange(v) }) }), (0,jsx_runtime.jsx)(Form_index_js_.FormGroup, { label: "Password", isRequired: true, fieldId: "password", children: (0,jsx_runtime.jsx)(TextInput_index_js_.TextInput, { id: "password", type: "password", value: props.password, onChange: (_e, v) => props.onPasswordChange(v) }) }), (0,jsx_runtime.jsx)(Form_index_js_.FormGroup, { fieldId: "create-secret", children: (0,jsx_runtime.jsx)(Checkbox_index_js_.Checkbox, { id: "create-secret", label: "Create a secret from these credentials (recommended)", isChecked: props.createSecret, onChange: (_e, v) => props.onCreateSecretChange(v), description: "Stores credentials securely as a Kubernetes secret for the operator to use" }) })] }))] })] }));

// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/layouts/Stack@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/layouts/Stack/index.js)
var Stack_index_js_ = __webpack_require__(4400);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Text@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Text/index.js)
var Text_index_js_ = __webpack_require__(208);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-icons/dist/dynamic/icons/plus-circle-icon@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-icons/dist/esm/icons/plus-circle-icon.js)
var plus_circle_icon_js_ = __webpack_require__(253);
;// ./src/app/hooks/useVSphereBrowse.ts


const API_BASE = '/api/proxy/plugin/vcf-migration-console/vcf-migration-api';
function useVSphereConnect() {
    const [loading, setLoading] = (0,consume_shared_module_default_react_17_0_singleton_.useState)(false);
    const [error, setError] = (0,consume_shared_module_default_react_17_0_singleton_.useState)(null);
    const connect = (0,consume_shared_module_default_react_17_0_singleton_.useCallback)(async (params) => {
        setLoading(true);
        setError(null);
        try {
            const res = await (0,dynamic_plugin_sdk_1_8_singleton_.consoleFetch)(`${API_BASE}/vsphere/connect`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(params),
            });
            const data = await res.json();
            if (!res.ok) {
                setError(data.error || res.statusText);
                return { datacenters: [], error: data.error || res.statusText };
            }
            return data;
        }
        catch (e) {
            const msg = e instanceof Error ? e.message : String(e);
            setError(msg);
            return { datacenters: [], error: msg };
        }
        finally {
            setLoading(false);
        }
    }, []);
    return { connect, loading, error };
}
function useVSphereList(endpoint, params) {
    const [items, setItems] = (0,consume_shared_module_default_react_17_0_singleton_.useState)([]);
    const [loading, setLoading] = (0,consume_shared_module_default_react_17_0_singleton_.useState)(false);
    const [error, setError] = (0,consume_shared_module_default_react_17_0_singleton_.useState)(null);
    const fetchList = (0,consume_shared_module_default_react_17_0_singleton_.useCallback)(async () => {
        if (!params.server || !params.datacenter)
            return;
        setLoading(true);
        setError(null);
        try {
            const searchParams = new URLSearchParams({
                server: params.server,
                datacenter: params.datacenter,
            });
            if (params.secretName) {
                searchParams.set('secretName', params.secretName);
                if (params.secretNamespace)
                    searchParams.set('secretNamespace', params.secretNamespace);
            }
            else if (params.username && params.password) {
                searchParams.set('username', params.username);
                searchParams.set('password', params.password);
            }
            const res = await (0,dynamic_plugin_sdk_1_8_singleton_.consoleFetch)(`${API_BASE}/vsphere/${endpoint}?${searchParams}`);
            const data = await res.json();
            if (!res.ok) {
                setError(data.error || res.statusText);
                setItems([]);
                return;
            }
            setItems(data.items || []);
        }
        catch (e) {
            setError(e instanceof Error ? e.message : String(e));
            setItems([]);
        }
        finally {
            setLoading(false);
        }
    }, [endpoint, params.server, params.datacenter, params.secretName, params.secretNamespace, params.username, params.password]);
    return { items, loading, error, fetchList };
}

// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Card@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Card/index.js)
var Card_index_js_ = __webpack_require__(1414);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-icons/dist/dynamic/icons/trash-icon@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-icons/dist/esm/icons/trash-icon.js)
var trash_icon_js_ = __webpack_require__(195);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Select@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Select/index.js)
var Select_index_js_ = __webpack_require__(2070);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/TextInputGroup@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/TextInputGroup/index.js)
var TextInputGroup_index_js_ = __webpack_require__(4490);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-icons/dist/dynamic/icons/times-icon@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-icons/dist/esm/icons/times-icon.js)
var times_icon_js_ = __webpack_require__(1607);
;// ./src/app/components/wizard/TypeaheadSelect.tsx












const TypeaheadSelect = ({ id, items, value, onChange, placeholder = 'Select...', isDisabled = false, isLoading = false, ...props }) => {
    const [isOpen, setIsOpen] = consume_shared_module_default_react_17_0_singleton_.useState(false);
    const [filterValue, setFilterValue] = consume_shared_module_default_react_17_0_singleton_.useState('');
    const textInputRef = consume_shared_module_default_react_17_0_singleton_.useRef(null);
    const filtered = consume_shared_module_default_react_17_0_singleton_.useMemo(() => {
        if (!filterValue)
            return items;
        const lower = filterValue.toLowerCase();
        return items.filter((item) => item.toLowerCase().includes(lower));
    }, [items, filterValue]);
    const handleSelect = consume_shared_module_default_react_17_0_singleton_.useCallback((_e, val) => {
        onChange(val);
        setIsOpen(false);
        setFilterValue('');
    }, [onChange]);
    const handleClear = consume_shared_module_default_react_17_0_singleton_.useCallback(() => {
        onChange('');
        setFilterValue('');
        textInputRef.current?.focus();
    }, [onChange]);
    const handleToggle = consume_shared_module_default_react_17_0_singleton_.useCallback(() => {
        if (!isDisabled)
            setIsOpen((prev) => !prev);
    }, [isDisabled]);
    const handleInputChange = consume_shared_module_default_react_17_0_singleton_.useCallback((_e, val) => {
        setFilterValue(val);
        if (!isOpen)
            setIsOpen(true);
    }, [isOpen]);
    const displayValue = isOpen ? filterValue : value;
    const toggle = consume_shared_module_default_react_17_0_singleton_.useCallback((toggleRef) => ((0,jsx_runtime.jsx)(MenuToggle_index_js_.MenuToggle, { ref: toggleRef, variant: "typeahead", onClick: handleToggle, isExpanded: isOpen, isDisabled: isDisabled, isFullWidth: true, children: (0,jsx_runtime.jsxs)(TextInputGroup_index_js_.TextInputGroup, { isPlain: true, children: [(0,jsx_runtime.jsx)(TextInputGroup_index_js_.TextInputGroupMain, { value: displayValue, onClick: handleToggle, onChange: handleInputChange, autoComplete: "off", innerRef: textInputRef, placeholder: placeholder, "aria-label": props['aria-label'] ?? placeholder, id: id }), (value || filterValue) && !isDisabled && ((0,jsx_runtime.jsxs)(TextInputGroup_index_js_.TextInputGroupUtilities, { children: [isLoading && (0,jsx_runtime.jsx)(Spinner_index_js_.Spinner, { size: "sm" }), (0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "plain", onClick: handleClear, "aria-label": "Clear", children: (0,jsx_runtime.jsx)(times_icon_js_.TimesIcon, {}) })] }))] }) })), [handleToggle, isOpen, isDisabled, displayValue, handleInputChange, placeholder, props, id, value, filterValue, isLoading, handleClear]);
    return ((0,jsx_runtime.jsx)(Select_index_js_.Select, { id: `${id}-select`, isOpen: isOpen, selected: value, onSelect: handleSelect, onOpenChange: setIsOpen, toggle: toggle, variant: "typeahead", isScrollable: true, maxMenuHeight: "300px", children: (0,jsx_runtime.jsx)(Select_index_js_.SelectList, { children: isLoading ? ((0,jsx_runtime.jsx)(Select_index_js_.SelectOption, { isDisabled: true, value: "loading", children: "Loading..." })) : filtered.length === 0 ? ((0,jsx_runtime.jsx)(Select_index_js_.SelectOption, { isDisabled: true, value: "no-results", children: "No results found" })) : (filtered.map((item) => ((0,jsx_runtime.jsx)(Select_index_js_.SelectOption, { value: item, children: item }, item)))) }) }));
};

;// ./src/app/components/wizard/FailureDomainEditor.tsx

















const sanitize = (s) => s.replace(/[^a-zA-Z0-9-]/g, '-').replace(/--+/g, '-').replace(/^-|-$/g, '').toLowerCase();
const lastSegment = (path) => {
    const parts = path.split('/').filter(Boolean);
    return parts[parts.length - 1] || path;
};
const FailureDomainEditor = ({ index, domain, onUpdate, onRemove, datacenters, server, username, password, secretRef, }) => {
    const prevDerivedName = consume_shared_module_default_react_17_0_singleton_.useRef('');
    const prevDerivedRegion = consume_shared_module_default_react_17_0_singleton_.useRef('');
    const prevDerivedZone = consume_shared_module_default_react_17_0_singleton_.useRef('');
    const [createFolder, setCreateFolder] = consume_shared_module_default_react_17_0_singleton_.useState(!domain.topology.folder);
    const listParams = consume_shared_module_default_react_17_0_singleton_.useMemo(() => ({
        server,
        datacenter: domain.topology.datacenter,
        secretName: secretRef?.name,
        secretNamespace: secretRef?.namespace,
        username,
        password,
    }), [server, domain.topology.datacenter, secretRef?.name, secretRef?.namespace, username, password]);
    const { items: clusters, loading: clustersLoading, fetchList: fetchClusters } = useVSphereList('clusters', listParams);
    const { items: datastores, loading: dsLoading, fetchList: fetchDatastores } = useVSphereList('datastores', listParams);
    const { items: networks, loading: netLoading, fetchList: fetchNetworks } = useVSphereList('networks', listParams);
    const { items: resourcePools, loading: rpLoading, fetchList: fetchResourcePools } = useVSphereList('resourcepools', listParams);
    const { items: templates, loading: tmplLoading, fetchList: fetchTemplates } = useVSphereList('templates', listParams);
    const { items: folders, loading: folderLoading, fetchList: fetchFolders } = useVSphereList('folders', listParams);
    consume_shared_module_default_react_17_0_singleton_.useEffect(() => {
        if (domain.topology.datacenter) {
            fetchClusters();
            fetchDatastores();
            fetchNetworks();
            fetchResourcePools();
            fetchTemplates();
            fetchFolders();
        }
    }, [domain.topology.datacenter, fetchClusters, fetchDatastores, fetchNetworks, fetchResourcePools, fetchTemplates, fetchFolders]);
    const handleDatacenterChange = consume_shared_module_default_react_17_0_singleton_.useCallback((dc) => {
        const dcName = lastSegment(dc);
        const derivedRegion = sanitize(dcName);
        const derivedName = `fd-${derivedRegion}`;
        const updates = {
            topology: { ...domain.topology, datacenter: dc, computeCluster: '', datastore: '', networks: [], template: '', folder: undefined, resourcePool: undefined },
        };
        if (!domain.name || domain.name === prevDerivedName.current) {
            updates.name = derivedName;
        }
        if (!domain.region || domain.region === prevDerivedRegion.current) {
            updates.region = derivedRegion;
        }
        if (!domain.zone || domain.zone === prevDerivedZone.current) {
            updates.zone = '';
        }
        prevDerivedName.current = derivedName;
        prevDerivedRegion.current = derivedRegion;
        onUpdate({ ...domain, ...updates });
    }, [domain, onUpdate]);
    const handleClusterChange = consume_shared_module_default_react_17_0_singleton_.useCallback((cluster) => {
        const clusterName = lastSegment(cluster);
        const derivedZone = sanitize(clusterName);
        const updates = {
            topology: { ...domain.topology, computeCluster: cluster },
        };
        if (!domain.zone || domain.zone === prevDerivedZone.current) {
            updates.zone = derivedZone;
        }
        prevDerivedZone.current = derivedZone;
        onUpdate({ ...domain, ...updates });
    }, [domain, onUpdate]);
    const updateTopology = consume_shared_module_default_react_17_0_singleton_.useCallback((field, value) => {
        onUpdate({
            ...domain,
            topology: { ...domain.topology, [field]: value },
        });
    }, [domain, onUpdate]);
    const handleCreateFolderChange = consume_shared_module_default_react_17_0_singleton_.useCallback((_e, checked) => {
        setCreateFolder(checked);
        if (checked) {
            updateTopology('folder', undefined);
        }
    }, [updateTopology]);
    const dcNotSelected = !domain.topology.datacenter;
    const idPrefix = `fd-${index}`;
    return ((0,jsx_runtime.jsxs)(Card_index_js_.Card, { isCompact: true, children: [(0,jsx_runtime.jsx)(Card_index_js_.CardHeader, { actions: {
                    actions: ((0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "plain", icon: (0,jsx_runtime.jsx)(trash_icon_js_.TrashIcon, {}), onClick: onRemove, "aria-label": "Remove failure domain" })),
                }, children: (0,jsx_runtime.jsxs)(Card_index_js_.CardTitle, { children: ["Failure domain ", index + 1, domain.name ? `: ${domain.name}` : ''] }) }), (0,jsx_runtime.jsx)(Card_index_js_.CardBody, { children: (0,jsx_runtime.jsxs)(Form_index_js_.Form, { children: [(0,jsx_runtime.jsx)(Form_index_js_.FormGroup, { label: "Datacenter", isRequired: true, fieldId: `${idPrefix}-dc`, children: (0,jsx_runtime.jsx)(TypeaheadSelect, { id: `${idPrefix}-dc`, items: datacenters, value: domain.topology.datacenter, onChange: handleDatacenterChange, placeholder: "Select datacenter", "aria-label": "Datacenter" }) }), (0,jsx_runtime.jsx)(Form_index_js_.FormGroup, { label: "Compute cluster", isRequired: true, fieldId: `${idPrefix}-cluster`, children: (0,jsx_runtime.jsx)(TypeaheadSelect, { id: `${idPrefix}-cluster`, items: clusters, value: domain.topology.computeCluster, onChange: handleClusterChange, placeholder: "Select cluster", isDisabled: dcNotSelected, isLoading: clustersLoading, "aria-label": "Compute cluster" }) }), (0,jsx_runtime.jsx)(Form_index_js_.FormGroup, { label: "Datastore", isRequired: true, fieldId: `${idPrefix}-ds`, children: (0,jsx_runtime.jsx)(TypeaheadSelect, { id: `${idPrefix}-ds`, items: datastores, value: domain.topology.datastore, onChange: (v) => updateTopology('datastore', v), placeholder: "Select datastore", isDisabled: dcNotSelected, isLoading: dsLoading, "aria-label": "Datastore" }) }), (0,jsx_runtime.jsx)(Form_index_js_.FormGroup, { label: "Network", isRequired: true, fieldId: `${idPrefix}-net`, children: (0,jsx_runtime.jsx)(TypeaheadSelect, { id: `${idPrefix}-net`, items: networks, value: domain.topology.networks?.[0] ?? '', onChange: (v) => updateTopology('networks', v ? [v] : []), placeholder: "Select network", isDisabled: dcNotSelected, isLoading: netLoading, "aria-label": "Network" }) }), (0,jsx_runtime.jsx)(Form_index_js_.FormGroup, { label: "Template (RHCOS)", isRequired: true, fieldId: `${idPrefix}-tmpl`, children: (0,jsx_runtime.jsx)(TypeaheadSelect, { id: `${idPrefix}-tmpl`, items: templates, value: domain.topology.template, onChange: (v) => updateTopology('template', v), placeholder: "Select template", isDisabled: dcNotSelected, isLoading: tmplLoading, "aria-label": "Template" }) }), (0,jsx_runtime.jsx)(Form_index_js_.FormGroup, { label: "Resource pool", fieldId: `${idPrefix}-rp`, children: (0,jsx_runtime.jsx)(TypeaheadSelect, { id: `${idPrefix}-rp`, items: resourcePools, value: domain.topology.resourcePool ?? '', onChange: (v) => updateTopology('resourcePool', v || undefined), placeholder: "Select resource pool (optional)", isDisabled: dcNotSelected, isLoading: rpLoading, "aria-label": "Resource pool" }) }), (0,jsx_runtime.jsxs)(Form_index_js_.FormGroup, { label: "Folder", fieldId: `${idPrefix}-folder`, children: [(0,jsx_runtime.jsx)(Checkbox_index_js_.Checkbox, { id: `${idPrefix}-create-folder`, label: "Create folder automatically", isChecked: createFolder, onChange: handleCreateFolderChange, className: "pf-v5-u-mb-sm" }), !createFolder && ((0,jsx_runtime.jsx)(TypeaheadSelect, { id: `${idPrefix}-folder`, items: folders, value: domain.topology.folder ?? '', onChange: (v) => updateTopology('folder', v || undefined), placeholder: "Select existing folder", isDisabled: dcNotSelected, isLoading: folderLoading, "aria-label": "Folder" })), (0,jsx_runtime.jsx)(Form_index_js_.FormHelperText, { children: (0,jsx_runtime.jsx)(HelperText_index_js_.HelperText, { children: (0,jsx_runtime.jsx)(HelperText_index_js_.HelperTextItem, { children: createFolder
                                                ? 'The operator will create a VM folder using the cluster infrastructure ID'
                                                : 'Select an existing VM folder on the target vCenter' }) }) })] }), (0,jsx_runtime.jsxs)(Form_index_js_.FormGroup, { label: "Name", isRequired: true, fieldId: `${idPrefix}-name`, children: [(0,jsx_runtime.jsx)(TextInput_index_js_.TextInput, { id: `${idPrefix}-name`, value: domain.name, onChange: (_e, v) => onUpdate({ ...domain, name: v }), placeholder: "fd-1" }), (0,jsx_runtime.jsx)(Form_index_js_.FormHelperText, { children: (0,jsx_runtime.jsx)(HelperText_index_js_.HelperText, { children: (0,jsx_runtime.jsx)(HelperText_index_js_.HelperTextItem, { children: "Auto-derived from datacenter; editable" }) }) })] }), (0,jsx_runtime.jsxs)(Form_index_js_.FormGroup, { label: "Region", isRequired: true, fieldId: `${idPrefix}-region`, children: [(0,jsx_runtime.jsx)(TextInput_index_js_.TextInput, { id: `${idPrefix}-region`, value: domain.region, onChange: (_e, v) => onUpdate({ ...domain, region: v }), placeholder: "region1" }), (0,jsx_runtime.jsx)(Form_index_js_.FormHelperText, { children: (0,jsx_runtime.jsx)(HelperText_index_js_.HelperText, { children: (0,jsx_runtime.jsx)(HelperText_index_js_.HelperTextItem, { children: "Auto-derived from datacenter; editable" }) }) })] }), (0,jsx_runtime.jsxs)(Form_index_js_.FormGroup, { label: "Zone", isRequired: true, fieldId: `${idPrefix}-zone`, children: [(0,jsx_runtime.jsx)(TextInput_index_js_.TextInput, { id: `${idPrefix}-zone`, value: domain.zone, onChange: (_e, v) => onUpdate({ ...domain, zone: v }), placeholder: "zone1" }), (0,jsx_runtime.jsx)(Form_index_js_.FormHelperText, { children: (0,jsx_runtime.jsx)(HelperText_index_js_.HelperText, { children: (0,jsx_runtime.jsx)(HelperText_index_js_.HelperTextItem, { children: "Auto-derived from compute cluster; editable" }) }) })] })] }) })] }));
};

;// ./src/app/components/wizard/FailureDomainStep.tsx





















const emptyFailureDomain = (server) => ({
    name: '',
    region: '',
    zone: '',
    server,
    topology: {
        datacenter: '',
        computeCluster: '',
        datastore: '',
        networks: [],
        template: '',
        folder: '',
    },
});
const FailureDomainStep = (props) => {
    const { connect, loading: connecting, error: connectError } = useVSphereConnect();
    const [datacenters, setDatacenters] = consume_shared_module_default_react_17_0_singleton_.useState([]);
    const [connected, setConnected] = consume_shared_module_default_react_17_0_singleton_.useState(false);
    const connectAttempted = consume_shared_module_default_react_17_0_singleton_.useRef(false);
    consume_shared_module_default_react_17_0_singleton_.useEffect(() => {
        if (connectAttempted.current || connected || !props.server)
            return;
        connectAttempted.current = true;
        connect({
            server: props.server,
            username: props.username,
            password: props.password,
            secretRef: props.secretRef,
        }).then((result) => {
            if (result?.datacenters?.length) {
                setDatacenters(result.datacenters);
                setConnected(true);
            }
        });
    }, [props.server, props.username, props.password, props.secretRef, connect, connected]);
    const handleRetry = consume_shared_module_default_react_17_0_singleton_.useCallback(() => {
        connectAttempted.current = false;
        setConnected(false);
        setDatacenters([]);
        connect({
            server: props.server,
            username: props.username,
            password: props.password,
            secretRef: props.secretRef,
        }).then((result) => {
            if (result?.datacenters?.length) {
                setDatacenters(result.datacenters);
                setConnected(true);
            }
        });
    }, [props.server, props.username, props.password, props.secretRef, connect]);
    const addDomain = () => {
        props.onFailureDomainsChange([
            ...props.failureDomains,
            emptyFailureDomain(props.server),
        ]);
    };
    const removeDomain = (index) => {
        props.onFailureDomainsChange(props.failureDomains.filter((_, i) => i !== index));
    };
    const updateDomain = (index, fd) => {
        const next = [...props.failureDomains];
        next[index] = fd;
        props.onFailureDomainsChange(next);
    };
    if (connecting) {
        return ((0,jsx_runtime.jsx)(Bullseye_index_js_.Bullseye, { children: (0,jsx_runtime.jsxs)(Stack_index_js_.Stack, { hasGutter: true, children: [(0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { isFilled: true, children: (0,jsx_runtime.jsx)(Bullseye_index_js_.Bullseye, { children: (0,jsx_runtime.jsx)(Spinner_index_js_.Spinner, { size: "xl", "aria-label": "Connecting to vCenter" }) }) }), (0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { children: (0,jsx_runtime.jsx)(Bullseye_index_js_.Bullseye, { children: (0,jsx_runtime.jsx)(Text_index_js_.TextContent, { children: (0,jsx_runtime.jsxs)(Text_index_js_.Text, { component: Text_index_js_.TextVariants.p, children: ["Connecting to ", props.server, "..."] }) }) }) })] }) }));
    }
    if (connectError && !connected) {
        return ((0,jsx_runtime.jsxs)(Stack_index_js_.Stack, { hasGutter: true, children: [(0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { children: (0,jsx_runtime.jsx)(Alert_index_js_.Alert, { variant: "danger", title: "Failed to connect to vCenter", isInline: true, children: connectError }) }), (0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { children: (0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "secondary", onClick: handleRetry, children: "Retry connection" }) })] }));
    }
    if (!connected) {
        return ((0,jsx_runtime.jsx)(Bullseye_index_js_.Bullseye, { children: (0,jsx_runtime.jsx)(Spinner_index_js_.Spinner, { size: "xl", "aria-label": "Connecting to vCenter" }) }));
    }
    if (props.failureDomains.length === 0) {
        return ((0,jsx_runtime.jsxs)(EmptyState_index_js_.EmptyState, { children: [(0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateHeader, { titleText: "No failure domains configured", headingLevel: "h4", icon: (0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateIcon, { icon: cubes_icon_js_.CubesIcon }) }), (0,jsx_runtime.jsxs)(EmptyState_index_js_.EmptyStateBody, { children: ["Connected to ", props.server, " (", datacenters.length, " datacenter", datacenters.length !== 1 ? 's' : '', " found). Add at least one failure domain to define the topology for the migration."] }), (0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateFooter, { children: (0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateActions, { children: (0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "primary", icon: (0,jsx_runtime.jsx)(plus_circle_icon_js_.PlusCircleIcon, {}), onClick: addDomain, children: "Add failure domain" }) }) })] }));
    }
    return ((0,jsx_runtime.jsxs)(Stack_index_js_.Stack, { hasGutter: true, children: [(0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { children: (0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "secondary", icon: (0,jsx_runtime.jsx)(plus_circle_icon_js_.PlusCircleIcon, {}), onClick: addDomain, children: "Add failure domain" }) }), props.failureDomains.map((fd, i) => ((0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { children: (0,jsx_runtime.jsx)(FailureDomainEditor, { index: i, domain: fd, onUpdate: (updated) => updateDomain(i, updated), onRemove: () => removeDomain(i), datacenters: datacenters, server: props.server, username: props.username, password: props.password, secretRef: props.secretRef }) }, i)))] }));
};

// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/DescriptionList@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/DescriptionList/index.js)
var DescriptionList_index_js_ = __webpack_require__(7472);
;// ./src/app/components/wizard/ReviewStep.tsx
















const ReviewStep = ({ migration }) => ((0,jsx_runtime.jsxs)(Stack_index_js_.Stack, { hasGutter: true, children: [(0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { children: (0,jsx_runtime.jsx)(Text_index_js_.TextContent, { children: (0,jsx_runtime.jsx)(Text_index_js_.Text, { component: Text_index_js_.TextVariants.p, children: "Review the migration configuration before creating the resource." }) }) }), (0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { children: (0,jsx_runtime.jsxs)(Card_index_js_.Card, { isPlain: true, isCompact: true, children: [(0,jsx_runtime.jsx)(Card_index_js_.CardTitle, { children: "General" }), (0,jsx_runtime.jsx)(Card_index_js_.CardBody, { children: (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionList, { isHorizontal: true, termWidth: "12ch", children: [(0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Name" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: migration.metadata.name })] }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Namespace" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: migration.metadata.namespace })] }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "State" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: (0,jsx_runtime.jsx)(Label_index_js_.Label, { color: "blue", children: migration.spec.state }) })] })] }) })] }) }), (0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { children: (0,jsx_runtime.jsx)(Divider_index_js_.Divider, {}) }), (0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { children: (0,jsx_runtime.jsxs)(Card_index_js_.Card, { isPlain: true, isCompact: true, children: [(0,jsx_runtime.jsx)(Card_index_js_.CardTitle, { children: "Credentials" }), (0,jsx_runtime.jsx)(Card_index_js_.CardBody, { children: (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionList, { isHorizontal: true, termWidth: "12ch", children: [(0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Secret name" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: migration.spec.targetVCenterCredentialsSecret.name || '(not set)' })] }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Secret namespace" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: migration.spec.targetVCenterCredentialsSecret.namespace || '(default)' })] })] }) })] }) }), migration.spec.failureDomains?.length > 0 && ((0,jsx_runtime.jsxs)(jsx_runtime.Fragment, { children: [(0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { children: (0,jsx_runtime.jsx)(Divider_index_js_.Divider, {}) }), (0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { children: (0,jsx_runtime.jsx)(Text_index_js_.TextContent, { children: (0,jsx_runtime.jsxs)(Text_index_js_.Text, { component: Text_index_js_.TextVariants.h3, children: ["Failure domains (", migration.spec.failureDomains.length, ")"] }) }) }), migration.spec.failureDomains.map((fd, i) => ((0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { children: (0,jsx_runtime.jsxs)(Card_index_js_.Card, { isCompact: true, children: [(0,jsx_runtime.jsx)(Card_index_js_.CardTitle, { children: fd.name || `Failure domain ${i + 1}` }), (0,jsx_runtime.jsx)(Card_index_js_.CardBody, { children: (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionList, { isHorizontal: true, isCompact: true, termWidth: "14ch", children: [(0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Server" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: fd.server || '-' })] }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Region" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: fd.region || '-' })] }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Zone" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: fd.zone || '-' })] }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Datacenter" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: fd.topology.datacenter || '-' })] }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Compute cluster" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: fd.topology.computeCluster || '-' })] }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Datastore" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: fd.topology.datastore || '-' })] }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Network" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: fd.topology.networks?.join(', ') || '-' })] }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Template" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: fd.topology.template || '-' })] }), fd.topology.folder && ((0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Folder" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: fd.topology.folder })] })), fd.topology.resourcePool && ((0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Resource pool" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: fd.topology.resourcePool })] }))] }) })] }) }, i)))] }))] }));

;// ./src/app/pages/MigrationWizard.tsx

















const CreateMigrationFooter = ({ onCreate, isCreating }) => {
    const { goToPrevStep, close } = (0,Wizard_index_js_.useWizardContext)();
    return ((0,jsx_runtime.jsxs)(Wizard_index_js_.WizardFooterWrapper, { children: [(0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "primary", onClick: onCreate, isLoading: isCreating, isDisabled: isCreating, children: "Create migration" }), (0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "secondary", onClick: goToPrevStep, isDisabled: isCreating, children: "Back" }), (0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "link", onClick: close, isDisabled: isCreating, children: "Cancel" })] }));
};
const MigrationWizard = () => {
    const history = (0,consume_shared_module_default_react_router_dom_5_singleton_.useHistory)();
    const [migrationName, setMigrationName] = consume_shared_module_default_react_17_0_singleton_.useState('');
    const [migrationNamespace, setMigrationNamespace] = consume_shared_module_default_react_17_0_singleton_.useState('openshift-vcf-migration');
    const [server, setServer] = consume_shared_module_default_react_17_0_singleton_.useState('');
    const [username, setUsername] = consume_shared_module_default_react_17_0_singleton_.useState('');
    const [password, setPassword] = consume_shared_module_default_react_17_0_singleton_.useState('');
    const [useSecretRef, setUseSecretRef] = consume_shared_module_default_react_17_0_singleton_.useState(false);
    const [secretName, setSecretName] = consume_shared_module_default_react_17_0_singleton_.useState('');
    const [secretNamespace, setSecretNamespace] = consume_shared_module_default_react_17_0_singleton_.useState('openshift-vcf-migration');
    const [createSecret, setCreateSecret] = consume_shared_module_default_react_17_0_singleton_.useState(false);
    const [failureDomains, setFailureDomains] = consume_shared_module_default_react_17_0_singleton_.useState([]);
    const [createError, setCreateError] = consume_shared_module_default_react_17_0_singleton_.useState(null);
    const [isCreating, setIsCreating] = consume_shared_module_default_react_17_0_singleton_.useState(false);
    const buildMigration = consume_shared_module_default_react_17_0_singleton_.useCallback(() => {
        const effectiveName = migrationName || 'vcf-migration';
        return {
            apiVersion: `${VmwareCloudFoundationMigrationModel.apiGroup}/${VmwareCloudFoundationMigrationModel.apiVersion}`,
            kind: VmwareCloudFoundationMigrationModel.kind,
            metadata: {
                name: effectiveName,
                namespace: migrationNamespace,
            },
            spec: {
                state: 'Pending',
                targetVCenterCredentialsSecret: {
                    name: useSecretRef ? secretName : (createSecret ? `${effectiveName}-vcenter-creds` : secretName),
                    namespace: useSecretRef ? secretNamespace : (createSecret ? migrationNamespace : secretNamespace),
                },
                failureDomains,
            },
        };
    }, [migrationName, migrationNamespace, useSecretRef, secretName, secretNamespace, createSecret, failureDomains]);
    const handleCreate = consume_shared_module_default_react_17_0_singleton_.useCallback(async () => {
        setCreateError(null);
        setIsCreating(true);
        const obj = buildMigration();
        try {
            if (createSecret && !useSecretRef && username && password && server) {
                const secretNs = obj.spec.targetVCenterCredentialsSecret.namespace || obj.metadata.namespace;
                const secretBody = {
                    apiVersion: 'v1',
                    kind: 'Secret',
                    metadata: {
                        name: obj.spec.targetVCenterCredentialsSecret.name,
                        namespace: secretNs,
                    },
                    type: 'Opaque',
                    stringData: {
                        [`${server}.username`]: username,
                        [`${server}.password`]: password,
                    },
                };
                const res = await (0,dynamic_plugin_sdk_1_8_singleton_.consoleFetch)(`/api/kubernetes/api/v1/namespaces/${secretNs}/secrets`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(secretBody),
                });
                if (!res.ok) {
                    const errData = await res.json().catch(() => ({}));
                    const msg = errData.message || res.statusText;
                    throw new Error(`Failed to create credentials secret: ${msg}`);
                }
            }
            await (0,dynamic_plugin_sdk_1_8_singleton_.k8sCreate)({
                model: VmwareCloudFoundationMigrationModel,
                data: obj,
            });
            history.push(`/vcf-migration/ns/${obj.metadata.namespace}/${obj.metadata.name}`);
        }
        catch (e) {
            setCreateError(e instanceof Error ? e.message : String(e));
        }
        finally {
            setIsCreating(false);
        }
    }, [buildMigration, history, createSecret, useSecretRef, username, password, server]);
    const handleClose = consume_shared_module_default_react_17_0_singleton_.useCallback(() => {
        history.push('/vcf-migration');
    }, [history]);
    return ((0,jsx_runtime.jsx)(index_js_.PageSection, { isFilled: true, padding: { default: 'noPadding' }, children: (0,jsx_runtime.jsxs)(Wizard_index_js_.Wizard, { header: (0,jsx_runtime.jsx)(Wizard_index_js_.WizardHeader, { title: "Create VCF Migration", description: "Configure target vCenter credentials and failure domains for the migration", onClose: handleClose, closeButtonAriaLabel: "Close wizard" }), onClose: handleClose, children: [(0,jsx_runtime.jsx)(Wizard_index_js_.WizardStep, { name: "Credentials", id: "credentials", children: (0,jsx_runtime.jsx)(CredentialsStep, { server: server, onServerChange: setServer, username: username, onUsernameChange: setUsername, password: password, onPasswordChange: setPassword, useSecretRef: useSecretRef, onUseSecretRefChange: setUseSecretRef, secretName: secretName, onSecretNameChange: setSecretName, secretNamespace: secretNamespace, onSecretNamespaceChange: setSecretNamespace, createSecret: createSecret, onCreateSecretChange: setCreateSecret, migrationName: migrationName, migrationNamespace: migrationNamespace, onMigrationNameChange: setMigrationName, onMigrationNamespaceChange: setMigrationNamespace }) }), (0,jsx_runtime.jsx)(Wizard_index_js_.WizardStep, { name: "Failure domains", id: "failure-domains", children: (0,jsx_runtime.jsx)(FailureDomainStep, { server: server, username: username, password: password, secretRef: useSecretRef ? { name: secretName, namespace: secretNamespace } : undefined, failureDomains: failureDomains, onFailureDomainsChange: setFailureDomains }) }), (0,jsx_runtime.jsxs)(Wizard_index_js_.WizardStep, { name: "Review", id: "review", footer: (0,jsx_runtime.jsx)(CreateMigrationFooter, { onCreate: handleCreate, isCreating: isCreating }), children: [createError && ((0,jsx_runtime.jsx)(Alert_index_js_.Alert, { variant: "danger", title: "Migration creation failed", isInline: true, className: "pf-v5-u-mb-md", children: createError })), (0,jsx_runtime.jsx)(ReviewStep, { migration: buildMigration() })] })] }) }));
};

// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Breadcrumb@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Breadcrumb/index.js)
var Breadcrumb_index_js_ = __webpack_require__(9592);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/layouts/Flex@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/layouts/Flex/index.js)
var Flex_index_js_ = __webpack_require__(6228);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/ProgressStepper@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/ProgressStepper/index.js)
var ProgressStepper_index_js_ = __webpack_require__(9396);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-core/dist/dynamic/components/Tabs@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-core/dist/esm/components/Tabs/index.js)
var Tabs_index_js_ = __webpack_require__(634);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-icons/dist/dynamic/icons/download-icon@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-icons/dist/esm/icons/download-icon.js)
var download_icon_js_ = __webpack_require__(6783);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-icons/dist/dynamic/icons/info-circle-icon@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-icons/dist/esm/icons/info-circle-icon.js)
var info_circle_icon_js_ = __webpack_require__(1769);
;// ./src/app/hooks/useMigrationEvents.ts

const MAX_EVENTS = 100;
const MAX_RETRIES = 5;
const BASE_RETRY_MS = 2000;
function useMigrationEvents(namespace, name) {
    const [events, setEvents] = (0,consume_shared_module_default_react_17_0_singleton_.useState)([]);
    const [error, setError] = (0,consume_shared_module_default_react_17_0_singleton_.useState)(null);
    const eventSourceRef = (0,consume_shared_module_default_react_17_0_singleton_.useRef)(null);
    const retriesRef = (0,consume_shared_module_default_react_17_0_singleton_.useRef)(0);
    const connect = (0,consume_shared_module_default_react_17_0_singleton_.useCallback)(() => {
        if (!namespace || !name)
            return;
        const url = `/api/proxy/plugin/vcf-migration-console/vcf-migration-api/events?namespace=${encodeURIComponent(namespace)}&name=${encodeURIComponent(name)}`;
        const es = new EventSource(url);
        eventSourceRef.current = es;
        es.onopen = () => {
            retriesRef.current = 0;
            setError(null);
        };
        es.onmessage = (e) => {
            try {
                const data = JSON.parse(e.data);
                setError(null);
                setEvents((prev) => {
                    const next = [...prev];
                    const idx = next.findIndex((ev) => ev.reason === data.reason &&
                        ev.message === data.message &&
                        ev.lastTimestamp === data.lastTimestamp);
                    if (idx >= 0) {
                        next[idx] = data;
                    }
                    else {
                        next.unshift(data);
                    }
                    return next.slice(0, MAX_EVENTS);
                });
            }
            catch {
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
    (0,consume_shared_module_default_react_17_0_singleton_.useEffect)(() => {
        setEvents([]);
        setError(null);
        retriesRef.current = 0;
        if (!namespace || !name)
            return;
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

;// ./src/app/components/EventStream.tsx
















const formatTimestamp = (ts) => {
    if (!ts)
        return '-';
    const d = new Date(ts);
    return d.toLocaleString();
};
const getEventColor = (type) => {
    switch (type) {
        case 'Normal':
            return 'blue';
        case 'Warning':
            return 'orange';
        default:
            return 'grey';
    }
};
const EventStream = (props) => {
    const { events, error } = useMigrationEvents(props.namespace, props.name);
    if (error) {
        return ((0,jsx_runtime.jsx)(Alert_index_js_.Alert, { variant: "warning", title: "Event stream unavailable", isInline: true, children: error }));
    }
    if (events.length === 0) {
        return ((0,jsx_runtime.jsxs)(EmptyState_index_js_.EmptyState, { children: [(0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateHeader, { titleText: "No events", headingLevel: "h4", icon: (0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateIcon, { icon: info_circle_icon_js_.InfoCircleIcon }) }), (0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateBody, { children: "No events have been recorded for this migration yet." })] }));
    }
    return ((0,jsx_runtime.jsxs)(Table_index_js_.Table, { "aria-label": "Migration events", variant: "compact", children: [(0,jsx_runtime.jsx)(Table_index_js_.Thead, { children: (0,jsx_runtime.jsxs)(Table_index_js_.Tr, { children: [(0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "Type" }), (0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "Reason" }), (0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "Message" }), (0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "Last seen" }), (0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "Count" })] }) }), (0,jsx_runtime.jsx)(Table_index_js_.Tbody, { children: events.map((ev, i) => ((0,jsx_runtime.jsxs)(Table_index_js_.Tr, { children: [(0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "Type", children: (0,jsx_runtime.jsx)(Label_index_js_.Label, { color: getEventColor(ev.type), isCompact: true, children: ev.type }) }), (0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "Reason", children: ev.reason }), (0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "Message", children: ev.message }), (0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "Last seen", children: formatTimestamp(ev.lastTimestamp) }), (0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "Count", children: ev.count })] }, `${ev.lastTimestamp}-${ev.reason}-${i}`))) })] }));
};

// EXTERNAL MODULE: ./node_modules/@patternfly/react-topology/dist/esm/index.js + 801 modules
var esm = __webpack_require__(2744);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-icons/dist/dynamic/icons/server-icon@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-icons/dist/esm/icons/server-icon.js)
var server_icon_js_ = __webpack_require__(3775);
// EXTERNAL MODULE: consume shared module (default) @patternfly/react-icons/dist/dynamic/icons/desktop-icon@^5.0.0 (strict) (fallback: ./node_modules/@patternfly/react-icons/dist/esm/icons/desktop-icon.js)
var desktop_icon_js_ = __webpack_require__(1231);
;// ./src/app/components/MachineTopologyGraph.tsx





/* ------------------------------------------------------------------ */
/* Constants                                                           */
/* ------------------------------------------------------------------ */
const NODE_DIAMETER = 75;
/** Badge colors by machine role. */
const BADGE_COLORS = {
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
const machineStatus = (phase, nodeReady) => {
    switch (phase) {
        case 'Provisioning':
        case 'Provisioned':
            return esm/* NodeStatus */.zIx.info;
        case 'Running':
            if (nodeReady === true)
                return esm/* NodeStatus */.zIx.success;
            if (nodeReady === false)
                return esm/* NodeStatus */.zIx.warning;
            return esm/* NodeStatus */.zIx.default;
        case 'Failed':
        case 'Deleting':
            return esm/* NodeStatus */.zIx.danger;
        default:
            return esm/* NodeStatus */.zIx.default;
    }
};
/* ------------------------------------------------------------------ */
/* Custom node                                                         */
/* ------------------------------------------------------------------ */
const CustomNode = ({ element }) => {
    const data = element.getData();
    const Icon = data.role === 'master' ? server_icon_js_.ServerIcon : desktop_icon_js_.DesktopIcon;
    const badge = BADGE_COLORS[data.role] ?? BADGE_COLORS.worker;
    return ((0,jsx_runtime.jsx)(esm/* DefaultNode */.Icc, { element: element, showStatusDecorator: true, badge: data.role === 'master' ? 'CP' : 'W', badgeColor: badge.bg, badgeTextColor: badge.text, badgeBorderColor: badge.border, children: (0,jsx_runtime.jsx)("g", { transform: "translate(25, 25)", children: (0,jsx_runtime.jsx)(Icon, { width: 25, height: 25 }) }) }));
};
/* ------------------------------------------------------------------ */
/* Factories                                                           */
/* ------------------------------------------------------------------ */
const layoutFactory = (_type, graph) => new esm/* ColaLayout */.CbB(graph, { layoutOnDrag: false });
const componentFactory = (kind, type) => {
    if (type === 'group')
        return esm/* DefaultGroup */.b7q;
    switch (kind) {
        case esm/* ModelKind */.g9r.graph:
            return esm/* GraphComponent */.uJG;
        case esm/* ModelKind */.g9r.node:
            return CustomNode; // eslint-disable-line @typescript-eslint/no-explicit-any
        case esm/* ModelKind */.g9r.edge:
            return esm/* DefaultEdge */.FjY;
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
const buildTopologyModel = (rows) => {
    const groups = new Map();
    const nodes = [];
    rows.forEach((row) => {
        const id = `machine-${row.machineName}`;
        const groupKey = row.vcenter || 'unknown';
        if (!groups.has(groupKey)) {
            groups.set(groupKey, []);
        }
        groups.get(groupKey).push(id);
        nodes.push({
            id,
            type: 'node',
            label: row.machineName,
            width: NODE_DIAMETER,
            height: NODE_DIAMETER,
            shape: row.role === 'master' ? esm/* NodeShape */.AG_.hexagon : esm/* NodeShape */.AG_.ellipse,
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
const MachineTopologyGraph = ({ rows }) => {
    const controller = consume_shared_module_default_react_17_0_singleton_.useMemo(() => {
        const viz = new esm/* Visualization */.Dib();
        viz.registerLayoutFactory(layoutFactory);
        viz.registerComponentFactory(componentFactory);
        viz.fromModel(buildTopologyModel(rows), false);
        return viz;
    }, [rows]);
    return ((0,jsx_runtime.jsx)("div", { style: { height: 500, border: '1px solid var(--pf-v5-global--BorderColor--100)' }, children: (0,jsx_runtime.jsx)(esm/* VisualizationProvider */.Uk6, { controller: controller, children: (0,jsx_runtime.jsx)(esm/* VisualizationSurface */.AUs, {}) }) }));
};

;// ./src/app/components/MachineTopologyView.tsx

























const machineAPINamespace = 'openshift-machine-api';
/* ------------------------------------------------------------------ */
/* Component                                                           */
/* ------------------------------------------------------------------ */
const MachineTopologyView = () => {
    /* K8s watches */
    const [machineSets, msLoaded] = (0,dynamic_plugin_sdk_1_8_singleton_.useK8sWatchResource)({
        groupVersionKind: {
            group: MachineSetModel.apiGroup,
            version: MachineSetModel.apiVersion,
            kind: MachineSetModel.kind,
        },
        namespace: machineAPINamespace,
        isList: true,
        namespaced: true,
    });
    const [machines, mLoaded] = (0,dynamic_plugin_sdk_1_8_singleton_.useK8sWatchResource)({
        groupVersionKind: {
            group: MachineModel.apiGroup,
            version: MachineModel.apiVersion,
            kind: MachineModel.kind,
        },
        namespace: machineAPINamespace,
        isList: true,
        namespaced: true,
    });
    const [nodes, nLoaded] = (0,dynamic_plugin_sdk_1_8_singleton_.useK8sWatchResource)({
        groupVersionKind: {
            group: NodeModel.apiGroup,
            version: NodeModel.apiVersion,
            kind: NodeModel.kind,
        },
        isList: true,
        namespaced: false,
    });
    const [cpmsList, cpmsLoaded] = (0,dynamic_plugin_sdk_1_8_singleton_.useK8sWatchResource)({
        groupVersionKind: {
            group: ControlPlaneMachineSetModel.apiGroup,
            version: ControlPlaneMachineSetModel.apiVersion,
            kind: ControlPlaneMachineSetModel.kind,
        },
        namespace: machineAPINamespace,
        isList: true,
        namespaced: true,
    });
    const loaded = msLoaded && mLoaded && nLoaded && cpmsLoaded;
    /* Build lookup maps */
    const nodeMap = consume_shared_module_default_react_17_0_singleton_.useMemo(() => {
        const m = {};
        nodes?.forEach((n) => { m[n.metadata.name] = n; });
        return m;
    }, [nodes]);
    const msNames = consume_shared_module_default_react_17_0_singleton_.useMemo(() => new Set(machineSets?.map((ms) => ms.metadata.name) ?? []), [machineSets]);
    /* Build rows */
    const rows = consume_shared_module_default_react_17_0_singleton_.useMemo(() => {
        if (!loaded)
            return [];
        const result = [];
        machines?.forEach((m) => {
            const msLabel = m.metadata.labels?.['machine.openshift.io/cluster-api-machineset'] ?? null;
            const roleLabel = m.metadata.labels?.['machine.openshift.io/cluster-api-machine-role'] ?? '';
            const nodeName = m.status?.nodeRef?.name ?? null;
            let nodeReady = null;
            if (nodeName) {
                const node = nodeMap[nodeName];
                const cond = node?.status?.conditions?.find((c) => c.type === 'Ready');
                nodeReady = cond?.status === 'True';
            }
            const role = roleLabel === 'master' ? 'master' : 'worker';
            const vcenter = m.spec?.providerSpec?.value?.workspace?.server ?? 'unknown';
            result.push({
                machineSetName: msLabel && msNames.has(msLabel) ? msLabel : null,
                machineName: m.metadata.name,
                machineNamespace: m.metadata.namespace,
                machinePhase: m.status?.phase ?? 'Unknown',
                nodeName,
                nodeReady,
                role,
                vcenter,
            });
        });
        result.sort((a, b) => a.machineName.localeCompare(b.machineName));
        return result;
    }, [loaded, machines, msNames, nodeMap]);
    const controlPlaneRows = consume_shared_module_default_react_17_0_singleton_.useMemo(() => rows.filter((r) => r.role === 'master'), [rows]);
    const workerRows = consume_shared_module_default_react_17_0_singleton_.useMemo(() => rows.filter((r) => r.role === 'worker'), [rows]);
    const cpms = cpmsList?.[0] ?? null;
    /* Loading */
    if (!loaded) {
        return ((0,jsx_runtime.jsx)(Bullseye_index_js_.Bullseye, { children: (0,jsx_runtime.jsx)(Spinner_index_js_.Spinner, { size: "lg", "aria-label": "Loading machine topology" }) }));
    }
    /* Empty */
    if (!machines?.length) {
        return ((0,jsx_runtime.jsxs)(EmptyState_index_js_.EmptyState, { children: [(0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateHeader, { titleText: "No machines found", headingLevel: "h4", icon: (0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateIcon, { icon: cubes_icon_js_.CubesIcon }) }), (0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateBody, { children: "No machines were found in the openshift-machine-api namespace." })] }));
    }
    return ((0,jsx_runtime.jsxs)(jsx_runtime.Fragment, { children: [(0,jsx_runtime.jsxs)(index_js_.PageSection, { variant: "light", className: "pf-v5-u-pb-lg", children: [(0,jsx_runtime.jsx)(Title_index_js_.Title, { headingLevel: "h2", className: "pf-v5-u-mb-md", children: "Control Plane" }), cpms && (0,jsx_runtime.jsx)(CPMSSummary, { cpms: cpms }), (0,jsx_runtime.jsx)(MachineTable, { rows: controlPlaneRows })] }), (0,jsx_runtime.jsxs)(index_js_.PageSection, { variant: "light", className: "pf-v5-u-pb-lg", children: [(0,jsx_runtime.jsx)(Title_index_js_.Title, { headingLevel: "h2", className: "pf-v5-u-mb-md", children: "Compute" }), (0,jsx_runtime.jsx)(MachineTable, { rows: workerRows })] }), (0,jsx_runtime.jsxs)(index_js_.PageSection, { variant: "light", children: [(0,jsx_runtime.jsx)(Title_index_js_.Title, { headingLevel: "h2", className: "pf-v5-u-mb-md", children: "Topology" }), (0,jsx_runtime.jsx)(MachineTopologyGraph, { rows: rows })] })] }));
};
/* ------------------------------------------------------------------ */
/* CPMS summary                                                        */
/* ------------------------------------------------------------------ */
const CPMSSummary = ({ cpms }) => ((0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionList, { isHorizontal: true, isCompact: true, className: "pf-v5-u-mb-md", children: [(0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "ControlPlaneMachineSet" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: (0,jsx_runtime.jsx)(dynamic_plugin_sdk_1_8_singleton_.ResourceLink, { groupVersionKind: {
                            group: ControlPlaneMachineSetModel.apiGroup,
                            version: ControlPlaneMachineSetModel.apiVersion,
                            kind: ControlPlaneMachineSetModel.kind,
                        }, name: cpms.metadata.name, namespace: cpms.metadata.namespace }) })] }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "State" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: (0,jsx_runtime.jsx)(Label_index_js_.Label, { color: cpms.spec.state === 'Active' ? 'green' : 'grey', isCompact: true, children: cpms.spec.state ?? 'Unknown' }) })] }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Replicas" }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListDescription, { children: [cpms.status?.readyReplicas ?? 0, " / ", cpms.spec.replicas ?? 0, " ready"] })] })] }));
/* ------------------------------------------------------------------ */
/* Shared machine table                                                */
/* ------------------------------------------------------------------ */
const MachineTable = ({ rows }) => {
    if (rows.length === 0) {
        return (0,jsx_runtime.jsx)(EmptyState_index_js_.EmptyStateBody, { children: "No machines in this category." });
    }
    return ((0,jsx_runtime.jsxs)(Table_index_js_.Table, { "aria-label": "Machines", variant: "compact", children: [(0,jsx_runtime.jsx)(Table_index_js_.Thead, { children: (0,jsx_runtime.jsxs)(Table_index_js_.Tr, { children: [(0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "MachineSet" }), (0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "Machine" }), (0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "Phase" }), (0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "Node" }), (0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "Status" }), (0,jsx_runtime.jsx)(Table_index_js_.Th, { children: "vCenter" })] }) }), (0,jsx_runtime.jsx)(Table_index_js_.Tbody, { children: rows.map((row) => ((0,jsx_runtime.jsxs)(Table_index_js_.Tr, { children: [(0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "MachineSet", children: row.machineSetName ? ((0,jsx_runtime.jsx)(dynamic_plugin_sdk_1_8_singleton_.ResourceLink, { groupVersionKind: {
                                    group: MachineSetModel.apiGroup,
                                    version: MachineSetModel.apiVersion,
                                    kind: MachineSetModel.kind,
                                }, name: row.machineSetName, namespace: machineAPINamespace })) : ((0,jsx_runtime.jsx)("span", { className: "pf-v5-u-color-200", children: "-" })) }), (0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "Machine", children: (0,jsx_runtime.jsx)(dynamic_plugin_sdk_1_8_singleton_.ResourceLink, { groupVersionKind: {
                                    group: MachineModel.apiGroup,
                                    version: MachineModel.apiVersion,
                                    kind: MachineModel.kind,
                                }, name: row.machineName, namespace: row.machineNamespace }) }), (0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "Phase", children: (0,jsx_runtime.jsx)(Label_index_js_.Label, { color: getPhaseColor(row.machinePhase), isCompact: true, children: row.machinePhase }) }), (0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "Node", children: row.nodeName ? ((0,jsx_runtime.jsx)(dynamic_plugin_sdk_1_8_singleton_.ResourceLink, { groupVersionKind: {
                                    group: NodeModel.apiGroup,
                                    version: NodeModel.apiVersion,
                                    kind: NodeModel.kind,
                                }, name: row.nodeName })) : ((0,jsx_runtime.jsx)("span", { className: "pf-v5-u-color-200", children: "-" })) }), (0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "Status", children: row.nodeReady === null ? ((0,jsx_runtime.jsx)(Label_index_js_.Label, { color: "grey", isCompact: true, children: "Pending" })) : row.nodeReady ? ((0,jsx_runtime.jsx)(Label_index_js_.Label, { color: "green", isCompact: true, children: "Ready" })) : ((0,jsx_runtime.jsx)(Label_index_js_.Label, { color: "red", isCompact: true, children: "Not Ready" })) }), (0,jsx_runtime.jsx)(Table_index_js_.Td, { dataLabel: "vCenter", children: row.vcenter !== 'unknown' ? row.vcenter : ((0,jsx_runtime.jsx)("span", { className: "pf-v5-u-color-200", children: "-" })) })] }, row.machineName))) })] }));
};
/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */
const getPhaseColor = (phase) => {
    switch (phase) {
        case 'Running':
            return 'green';
        case 'Provisioning':
        case 'Provisioned':
            return 'blue';
        case 'Failed':
        case 'Deleting':
            return 'red';
        default:
            return 'grey';
    }
};

;// ./src/app/components/MigrationLogs.tsx














const OPERATOR_NAMESPACE = 'openshift-vcf-migration';
const TAIL_LINES = 500;
const REFRESH_INTERVAL_MS = 10000;
const MigrationLogs = () => {
    const [pods, setPods] = consume_shared_module_default_react_17_0_singleton_.useState([]);
    const [selectedPod, setSelectedPod] = consume_shared_module_default_react_17_0_singleton_.useState('');
    const [podSelectOpen, setPodSelectOpen] = consume_shared_module_default_react_17_0_singleton_.useState(false);
    const [logs, setLogs] = consume_shared_module_default_react_17_0_singleton_.useState('');
    const [loading, setLoading] = consume_shared_module_default_react_17_0_singleton_.useState(true);
    const [error, setError] = consume_shared_module_default_react_17_0_singleton_.useState(null);
    const logRef = consume_shared_module_default_react_17_0_singleton_.useRef(null);
    const intervalRef = consume_shared_module_default_react_17_0_singleton_.useRef(null);
    consume_shared_module_default_react_17_0_singleton_.useEffect(() => {
        let cancelled = false;
        const fetchPods = async () => {
            try {
                const data = (await (0,dynamic_plugin_sdk_1_8_singleton_.consoleFetchJSON)(`/api/kubernetes/api/v1/namespaces/${OPERATOR_NAMESPACE}/pods?labelSelector=control-plane%3Dcontroller-manager`));
                const names = data.items?.map((p) => p.metadata.name) ?? [];
                if (!cancelled) {
                    setPods(names);
                    if (names.length > 0 && !selectedPod)
                        setSelectedPod(names[0]);
                }
            }
            catch (e) {
                if (!cancelled)
                    setError(`Failed to list operator pods: ${e instanceof Error ? e.message : String(e)}`);
            }
        };
        fetchPods();
        return () => { cancelled = true; };
    }, []);
    const fetchLogs = consume_shared_module_default_react_17_0_singleton_.useCallback(async (podName) => {
        if (!podName)
            return;
        try {
            const raw = await (0,dynamic_plugin_sdk_1_8_singleton_.consoleFetchText)(`/api/kubernetes/api/v1/namespaces/${OPERATOR_NAMESPACE}/pods/${podName}/log?tailLines=${TAIL_LINES}&container=manager`);
            setLogs(raw);
            setError(null);
        }
        catch (e) {
            setError(`Failed to fetch logs: ${e instanceof Error ? e.message : String(e)}`);
        }
        finally {
            setLoading(false);
        }
    }, []);
    consume_shared_module_default_react_17_0_singleton_.useEffect(() => {
        if (!selectedPod)
            return;
        setLoading(true);
        fetchLogs(selectedPod);
        intervalRef.current = setInterval(() => fetchLogs(selectedPod), REFRESH_INTERVAL_MS);
        return () => {
            if (intervalRef.current)
                clearInterval(intervalRef.current);
        };
    }, [selectedPod, fetchLogs]);
    consume_shared_module_default_react_17_0_singleton_.useEffect(() => {
        if (logRef.current) {
            logRef.current.scrollTop = logRef.current.scrollHeight;
        }
    }, [logs]);
    if (error && !logs) {
        return (0,jsx_runtime.jsx)(Alert_index_js_.Alert, { variant: "warning", title: "Logs unavailable", isInline: true, children: error });
    }
    return ((0,jsx_runtime.jsxs)("div", { style: { display: 'flex', flexDirection: 'column', height: '100%' }, children: [(0,jsx_runtime.jsx)(Toolbar_index_js_.Toolbar, { children: (0,jsx_runtime.jsxs)(Toolbar_index_js_.ToolbarContent, { children: [(0,jsx_runtime.jsx)(Toolbar_index_js_.ToolbarItem, { children: (0,jsx_runtime.jsx)(Select_index_js_.Select, { isOpen: podSelectOpen, selected: selectedPod, onSelect: (_e, val) => { setSelectedPod(val); setPodSelectOpen(false); }, onOpenChange: setPodSelectOpen, toggle: (toggleRef) => ((0,jsx_runtime.jsx)(MenuToggle_index_js_.MenuToggle, { ref: toggleRef, onClick: () => setPodSelectOpen((p) => !p), isExpanded: podSelectOpen, style: { minWidth: 300 }, children: selectedPod || 'Select pod' })), children: (0,jsx_runtime.jsx)(Select_index_js_.SelectList, { children: pods.map((p) => ((0,jsx_runtime.jsx)(Select_index_js_.SelectOption, { value: p, children: p }, p))) }) }) }), (0,jsx_runtime.jsx)(Toolbar_index_js_.ToolbarItem, { children: (0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "plain", onClick: () => { if (selectedPod)
                                    fetchLogs(selectedPod); }, children: "Refresh" }) })] }) }), loading ? ((0,jsx_runtime.jsx)(Bullseye_index_js_.Bullseye, { children: (0,jsx_runtime.jsx)(Spinner_index_js_.Spinner, { size: "lg", "aria-label": "Loading logs" }) })) : ((0,jsx_runtime.jsx)("pre", { ref: logRef, style: {
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
                }, children: logs || 'No logs available.' }))] }));
};

;// ./src/app/pages/MigrationDetailPage.tsx







































const MigrationDetailPage_migrationGVK = {
    group: 'migration.openshift.io',
    version: 'v1alpha1',
    kind: 'VmwareCloudFoundationMigration',
};
const conditionOrder = [
    'InfrastructurePrepared',
    'DestinationInitialized',
    'MultiSiteConfigured',
    'WorkloadMigrated',
    'SourceCleaned',
    'Ready',
];
const conditionLabels = {
    InfrastructurePrepared: 'Infrastructure prepared',
    DestinationInitialized: 'Destination initialized',
    MultiSiteConfigured: 'Multi-site configured',
    WorkloadMigrated: 'Workload migrated',
    SourceCleaned: 'Source cleaned',
    Ready: 'Ready',
};
const MigrationDetailPage_migrationStates = ['Pending', 'Running', 'Paused'];
const MigrationDetailPage_getStateColor = (state) => {
    switch (state) {
        case 'Running':
            return 'blue';
        case 'Paused':
            return 'orange';
        default:
            return 'grey';
    }
};
const MigrationDetailPage = () => {
    const params = (0,consume_shared_module_default_react_router_dom_5_singleton_.useParams)();
    const location = (0,consume_shared_module_default_react_router_dom_5_singleton_.useLocation)();
    const history = (0,consume_shared_module_default_react_router_dom_5_singleton_.useHistory)();
    const { ns, name } = consume_shared_module_default_react_17_0_singleton_.useMemo(() => {
        if (params.ns && params.name)
            return params;
        const match = location.pathname.match(/\/vcf-migration\/ns\/([^/]+)\/([^/]+)/);
        if (match)
            return { ns: match[1], name: match[2] };
        return { ns: '', name: '' };
    }, [params, location.pathname]);
    const watchSpec = ns && name
        ? {
            groupVersionKind: MigrationDetailPage_migrationGVK,
            name,
            namespace: ns,
            namespaced: true,
            isList: false,
        }
        : null;
    const [migration, loaded, loadError] = (0,dynamic_plugin_sdk_1_8_singleton_.useK8sWatchResource)(watchSpec);
    const [actionsOpen, setActionsOpen] = consume_shared_module_default_react_17_0_singleton_.useState(false);
    const [actionError, setActionError] = consume_shared_module_default_react_17_0_singleton_.useState(null);
    const [activeTab, setActiveTab] = consume_shared_module_default_react_17_0_singleton_.useState('details');
    const handleSetState = consume_shared_module_default_react_17_0_singleton_.useCallback(async (state) => {
        setActionsOpen(false);
        if (!migration)
            return;
        try {
            await (0,dynamic_plugin_sdk_1_8_singleton_.k8sPatch)({
                model: VmwareCloudFoundationMigrationModel,
                resource: migration,
                data: [{ op: 'replace', path: '/spec/state', value: state }],
            });
        }
        catch (e) {
            setActionError(`Failed to set state to ${state}: ${e instanceof Error ? e.message : String(e)}`);
        }
    }, [migration]);
    const handleDelete = consume_shared_module_default_react_17_0_singleton_.useCallback(async () => {
        setActionsOpen(false);
        if (!migration)
            return;
        try {
            await (0,dynamic_plugin_sdk_1_8_singleton_.k8sDelete)({
                model: VmwareCloudFoundationMigrationModel,
                resource: migration,
            });
            history.push('/vcf-migration');
        }
        catch (e) {
            setActionError(`Failed to delete: ${e instanceof Error ? e.message : String(e)}`);
        }
    }, [migration, history]);
    const handleDownloadMetadata = consume_shared_module_default_react_17_0_singleton_.useCallback(async () => {
        if (!ns || !name)
            return;
        try {
            const url = `/api/proxy/plugin/vcf-migration-console/vcf-migration-api/metadata?namespace=${encodeURIComponent(ns)}&name=${encodeURIComponent(name)}`;
            const response = await (0,dynamic_plugin_sdk_1_8_singleton_.consoleFetch)(url);
            const blob = await response.blob();
            const link = document.createElement('a');
            link.href = URL.createObjectURL(blob);
            link.download = `${name}-metadata.json`;
            link.click();
            URL.revokeObjectURL(link.href);
        }
        catch (e) {
            setActionError(`Failed to download metadata: ${e instanceof Error ? e.message : String(e)}`);
        }
    }, [ns, name]);
    if (!ns || !name) {
        return ((0,jsx_runtime.jsxs)(index_js_.PageSection, { children: [(0,jsx_runtime.jsx)(Title_index_js_.Title, { headingLevel: "h1", children: "Migration not found" }), (0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "link", onClick: () => history.push('/vcf-migration'), children: "Back to list" })] }));
    }
    if (loadError) {
        return ((0,jsx_runtime.jsxs)(index_js_.PageSection, { children: [(0,jsx_runtime.jsx)(Alert_index_js_.Alert, { variant: "danger", title: "Error loading migration", isInline: true, children: String(loadError) }), (0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "link", onClick: () => history.push('/vcf-migration'), className: "pf-v5-u-mt-md", children: "Back to list" })] }));
    }
    if (!loaded || !migration) {
        return ((0,jsx_runtime.jsx)(Bullseye_index_js_.Bullseye, { children: (0,jsx_runtime.jsx)(Spinner_index_js_.Spinner, { size: "xl", "aria-label": "Loading migration" }) }));
    }
    const getCondition = (type) => migration.status?.conditions?.find((c) => c.type === type);
    const isConditionTrue = (type) => getCondition(type)?.status === 'True';
    return ((0,jsx_runtime.jsxs)(jsx_runtime.Fragment, { children: [(0,jsx_runtime.jsxs)(index_js_.PageSection, { variant: "light", className: "pf-v5-u-pb-0", children: [(0,jsx_runtime.jsxs)(Breadcrumb_index_js_.Breadcrumb, { className: "pf-v5-u-mb-sm", children: [(0,jsx_runtime.jsx)(Breadcrumb_index_js_.BreadcrumbItem, { onClick: () => history.push('/vcf-migration'), children: "VCF Migrations" }), (0,jsx_runtime.jsx)(Breadcrumb_index_js_.BreadcrumbItem, { isActive: true, children: "Migration details" })] }), (0,jsx_runtime.jsxs)(Flex_index_js_.Flex, { alignItems: { default: 'alignItemsCenter' }, children: [(0,jsx_runtime.jsx)(Flex_index_js_.FlexItem, { children: (0,jsx_runtime.jsx)(Title_index_js_.Title, { headingLevel: "h1", className: "pf-v5-u-mr-sm", style: { display: 'inline' }, children: migration.metadata.name }) }), (0,jsx_runtime.jsx)(Flex_index_js_.FlexItem, { children: (0,jsx_runtime.jsx)(Label_index_js_.Label, { color: MigrationDetailPage_getStateColor(migration.spec.state), isCompact: true, children: migration.spec.state }) }), (0,jsx_runtime.jsx)(Flex_index_js_.FlexItem, { align: { default: 'alignRight' }, children: (0,jsx_runtime.jsx)(Dropdown_index_js_.Dropdown, { isOpen: actionsOpen, onSelect: () => setActionsOpen(false), onOpenChange: setActionsOpen, toggle: (toggleRef) => ((0,jsx_runtime.jsx)(MenuToggle_index_js_.MenuToggle, { ref: toggleRef, variant: "primary", onClick: () => setActionsOpen((prev) => !prev), isExpanded: actionsOpen, children: "Actions" })), popperProps: { position: 'right' }, children: (0,jsx_runtime.jsxs)(Dropdown_index_js_.DropdownList, { children: [MigrationDetailPage_migrationStates.map((state) => ((0,jsx_runtime.jsxs)(Dropdown_index_js_.DropdownItem, { onClick: () => handleSetState(state), isDisabled: migration.spec.state === state, description: migration.spec.state === state ? 'Current state' : undefined, children: ["Set ", state] }, state))), (0,jsx_runtime.jsx)(Divider_index_js_.Divider, {}), (0,jsx_runtime.jsx)(Dropdown_index_js_.DropdownItem, { onClick: handleDelete, isDanger: true, children: "Delete" }, "delete")] }) }) })] }), actionError && ((0,jsx_runtime.jsx)(Alert_index_js_.Alert, { variant: "danger", title: "Action failed", isInline: true, className: "pf-v5-u-mt-sm", actionClose: (0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "plain", onClick: () => setActionError(null), children: "Dismiss" }), children: actionError })), (0,jsx_runtime.jsxs)(Tabs_index_js_.Tabs, { activeKey: activeTab, onSelect: (_e, key) => setActiveTab(key), className: "pf-v5-u-mt-md", style: { marginBottom: -1 }, children: [(0,jsx_runtime.jsx)(Tabs_index_js_.Tab, { eventKey: "details", title: (0,jsx_runtime.jsx)(Tabs_index_js_.TabTitleText, { children: "Details" }) }), (0,jsx_runtime.jsx)(Tabs_index_js_.Tab, { eventKey: "yaml", title: (0,jsx_runtime.jsx)(Tabs_index_js_.TabTitleText, { children: "YAML" }) }), (0,jsx_runtime.jsx)(Tabs_index_js_.Tab, { eventKey: "machines", title: (0,jsx_runtime.jsx)(Tabs_index_js_.TabTitleText, { children: "Machines" }) }), (0,jsx_runtime.jsx)(Tabs_index_js_.Tab, { eventKey: "logs", title: (0,jsx_runtime.jsx)(Tabs_index_js_.TabTitleText, { children: "Logs" }) }), (0,jsx_runtime.jsx)(Tabs_index_js_.Tab, { eventKey: "events", title: (0,jsx_runtime.jsx)(Tabs_index_js_.TabTitleText, { children: "Events" }) })] })] }), activeTab === 'details' && ((0,jsx_runtime.jsx)(index_js_.PageSection, { children: (0,jsx_runtime.jsxs)(Stack_index_js_.Stack, { hasGutter: true, children: [(0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { children: (0,jsx_runtime.jsxs)(Card_index_js_.Card, { children: [(0,jsx_runtime.jsx)(Card_index_js_.CardTitle, { children: "Overview" }), (0,jsx_runtime.jsx)(Card_index_js_.CardBody, { children: (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionList, { isHorizontal: true, columnModifier: { default: '2Col' }, children: [(0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Namespace" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: migration.metadata.namespace })] }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "State" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: (0,jsx_runtime.jsx)(Label_index_js_.Label, { color: MigrationDetailPage_getStateColor(migration.spec.state), children: migration.spec.state }) })] }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Start time" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: migration.status?.startTime
                                                                ? new Date(migration.status.startTime).toLocaleString()
                                                                : '-' })] }), (0,jsx_runtime.jsxs)(DescriptionList_index_js_.DescriptionListGroup, { children: [(0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListTerm, { children: "Completion time" }), (0,jsx_runtime.jsx)(DescriptionList_index_js_.DescriptionListDescription, { children: migration.status?.completionTime
                                                                ? new Date(migration.status.completionTime).toLocaleString()
                                                                : '-' })] })] }) })] }) }), isConditionTrue('SourceCleaned') && ((0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { children: (0,jsx_runtime.jsxs)(Card_index_js_.Card, { children: [(0,jsx_runtime.jsx)(Card_index_js_.CardTitle, { children: "Installer metadata" }), (0,jsx_runtime.jsxs)(Card_index_js_.CardBody, { children: [(0,jsx_runtime.jsx)(Button_index_js_.Button, { variant: "secondary", icon: (0,jsx_runtime.jsx)(download_icon_js_.DownloadIcon, {}), onClick: handleDownloadMetadata, children: "Download metadata.json" }), (0,jsx_runtime.jsxs)("p", { className: "pf-v5-u-mt-sm pf-v5-u-color-200", style: { fontSize: 'var(--pf-v5-global--FontSize--sm)' }, children: ["Replacement installer metadata with destination vCenter configuration. Use this file to destroy the cluster with ", (0,jsx_runtime.jsx)("code", { children: "openshift-install destroy cluster" }), "."] })] })] }) })), (0,jsx_runtime.jsx)(Stack_index_js_.StackItem, { children: (0,jsx_runtime.jsxs)(Card_index_js_.Card, { children: [(0,jsx_runtime.jsx)(Card_index_js_.CardTitle, { children: "Migration progress" }), (0,jsx_runtime.jsx)(Card_index_js_.CardBody, { children: (0,jsx_runtime.jsx)(ProgressStepper_index_js_.ProgressStepper, { isVertical: true, children: conditionOrder.map((type) => {
                                                const cond = getCondition(type);
                                                const isDone = isConditionTrue(type);
                                                const isCurrent = cond?.status === 'False' &&
                                                    cond?.reason !== 'Failed';
                                                let variant = 'default';
                                                if (isDone)
                                                    variant = 'success';
                                                else if (cond?.reason === 'Failed')
                                                    variant = 'danger';
                                                else if (isCurrent)
                                                    variant = 'info';
                                                return ((0,jsx_runtime.jsx)(ProgressStepper_index_js_.ProgressStep, { variant: variant, id: type, titleId: `${type}-title`, "aria-label": conditionLabels[type] || type, description: cond?.message ?? (isDone ? 'Complete' : 'Pending'), children: conditionLabels[type] || type }, type));
                                            }) }) })] }) })] }) })), activeTab === 'yaml' && ((0,jsx_runtime.jsx)("div", { style: { display: 'flex', flex: 1, flexDirection: 'column', height: 'calc(100vh - 250px)', minHeight: 400 }, children: (0,jsx_runtime.jsx)(consume_shared_module_default_react_17_0_singleton_.Suspense, { fallback: (0,jsx_runtime.jsx)(Bullseye_index_js_.Bullseye, { children: (0,jsx_runtime.jsx)(Spinner_index_js_.Spinner, { size: "xl", "aria-label": "Loading editor" }) }), children: (0,jsx_runtime.jsx)(dynamic_plugin_sdk_1_8_singleton_.ResourceYAMLEditor, { initialResource: migration }) }) })), activeTab === 'machines' && ((0,jsx_runtime.jsx)(index_js_.PageSection, { children: (0,jsx_runtime.jsx)(MachineTopologyView, { namespace: ns }) })), activeTab === 'logs' && ((0,jsx_runtime.jsx)("div", { style: { display: 'flex', flexDirection: 'column', height: 'calc(100vh - 250px)', minHeight: 400 }, children: (0,jsx_runtime.jsx)(MigrationLogs, {}) })), activeTab === 'events' && ((0,jsx_runtime.jsx)(index_js_.PageSection, { children: (0,jsx_runtime.jsx)(EventStream, { namespace: ns, name: name }) }))] }));
};

;// ./src/app/index.ts






/***/ },

/***/ 1069
() {

/* (ignored) */

/***/ }

}]);
//# sourceMappingURL=exposed-migrationPlugin.chunk.js.map