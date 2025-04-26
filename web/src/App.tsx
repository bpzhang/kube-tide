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
import StatefulSets from './pages/workloads/StatefulSets';
import StatefulSetDetailPage from './pages/workloads/StatefulSetDetailPage';
import Dashboard from './pages/Dashboard';

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        {/* 独立页面路由 - 不使用MainLayout */}
        <Route path="/workloads/pods/:clusterName/:namespace/:podName/logs" element={<PodLogsPage />} />
        <Route path="/workloads/pods/:clusterName/:namespace/:podName/terminal" element={<PodTerminalPage />} />
        
        {/* 主布局路由 */}
        <Route path="/" element={<MainLayout />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="clusters">
            <Route index element={<Clusters />} />
            <Route path=":clusterName" element={<ClusterDetail />} />
          </Route>
          <Route path="nodes" element={<Nodes />} />
          <Route path="nodes/:clusterName/:nodeName" element={<NodeDetail />} />
          <Route path="workloads">
            <Route path="pods" element={<Pods />} />
            <Route path="pods/:clusterName/:namespace/:podName" element={<PodDetailPage />} />
            <Route path="services" element={<Services />} />
            <Route path="deployments" element={<Deployments />} />
            <Route path="deployments/:namespace" element={<Deployments />} />
            <Route path="statefulsets" element={<StatefulSets />} />
            <Route path="statefulsets/detail/:clusterName/:namespace/:statefulsetName" element={<StatefulSetDetailPage />} />
          </Route>
          <Route path="clusters/:clusterName/deployments" element={<Deployments />} />
          <Route path="clusters/:clusterName/deployments/:namespace" element={<Deployments />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
};

export default App;