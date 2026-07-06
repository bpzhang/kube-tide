import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import MainLayout from './layouts/MainLayout';
import Clusters from './pages/Clusters';
import ClusterDetail from './pages/ClusterDetail';
import Nodes from './pages/Nodes';
import NodeDetail from './pages/NodeDetail';
import Pods from './pages/workloads/Pods';
import PodDetailPage from './pages/workloads/PodDetailPage';
import PodLogsPage from './pages/workloads/PodLogsPage';
import PodTerminalPage from './pages/workloads/PodTerminalPage';
import Services from './pages/workloads/Services';
import Deployments from './pages/workloads/Deployments';
import DeploymentDetailPage from './pages/workloads/DeploymentDetailPage';
import StatefulSets from './pages/workloads/StatefulSets';
import StatefulSetDetailPage from './pages/workloads/StatefulSetDetailPage';
import HPAs from './pages/workloads/HPAs';
import DaemonSets from './pages/workloads/DaemonSets';
import Jobs from './pages/workloads/Jobs';
import CronJobs from './pages/workloads/CronJobs';
import ConfigMaps from './pages/config/ConfigMaps';
import Secrets from './pages/config/Secrets';
import Namespaces from './pages/Namespaces';
import Ingresses from './pages/network/Ingresses';
import NetworkPolicies from './pages/network/NetworkPolicies';
import PVCs from './pages/storage/PVCs';
import PVs from './pages/storage/PVs';
import StorageClasses from './pages/storage/StorageClasses';
import ResourceQuotas from './pages/governance/ResourceQuotas';
import LimitRanges from './pages/governance/LimitRanges';
import PDBs from './pages/governance/PDBs';
import RBAC from './pages/governance/RBAC';
import LabelLogs from './pages/observability/LabelLogs';
import Dashboard from './pages/Dashboard';

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/workloads/pods/:clusterName/:namespace/:podName/logs" element={<PodLogsPage />} />
        <Route path="/workloads/pods/:clusterName/:namespace/:podName/terminal" element={<PodTerminalPage />} />

        <Route path="/" element={<MainLayout />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="clusters">
            <Route index element={<Clusters />} />
            <Route path=":clusterName" element={<ClusterDetail />} />
          </Route>
          <Route path="nodes" element={<Nodes />} />
          <Route path="nodes/:clusterName/:nodeName" element={<NodeDetail />} />
          <Route path="namespaces" element={<Namespaces />} />
          <Route path="config">
            <Route path="configmaps" element={<ConfigMaps />} />
            <Route path="secrets" element={<Secrets />} />
          </Route>
          <Route path="workloads">
            <Route path="pods" element={<Pods />} />
            <Route path="pods/:clusterName/:namespace/:podName" element={<PodDetailPage />} />
            <Route path="services" element={<Services />} />
            <Route path="deployments" element={<Deployments />} />
            <Route path="deployments/detail/:clusterName/:namespace/:deploymentName" element={<DeploymentDetailPage />} />
            <Route path="deployments/ns/:namespace" element={<Deployments />} />
            <Route path="statefulsets" element={<StatefulSets />} />
            <Route path="statefulsets/detail/:clusterName/:namespace/:statefulsetName" element={<StatefulSetDetailPage />} />
            <Route path="hpas" element={<HPAs />} />
            <Route path="daemonsets" element={<DaemonSets />} />
            <Route path="jobs" element={<Jobs />} />
            <Route path="cronjobs" element={<CronJobs />} />
          </Route>
          <Route path="network">
            <Route path="ingresses" element={<Ingresses />} />
            <Route path="networkpolicies" element={<NetworkPolicies />} />
          </Route>
          <Route path="storage">
            <Route path="pvcs" element={<PVCs />} />
            <Route path="pvs" element={<PVs />} />
            <Route path="storageclasses" element={<StorageClasses />} />
          </Route>
          <Route path="governance">
            <Route path="resourcequotas" element={<ResourceQuotas />} />
            <Route path="limitranges" element={<LimitRanges />} />
            <Route path="pdbs" element={<PDBs />} />
            <Route path="rbac" element={<RBAC />} />
          </Route>
          <Route path="observability">
            <Route path="label-logs" element={<LabelLogs />} />
          </Route>
          <Route path="clusters/:clusterName/deployments" element={<Deployments />} />
          <Route path="clusters/:clusterName/deployments/:namespace" element={<Deployments />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
};

export default App;
