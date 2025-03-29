import React, { useEffect, useRef, useState } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { AttachAddon } from '@xterm/addon-attach';
import '@xterm/xterm/css/xterm.css';
import { Card, Alert, Spin, Button, Space, message } from 'antd';
import { ReloadOutlined, BugOutlined } from '@ant-design/icons';
import { checkPodExists } from '@/api/pod'; // 导入检查Pod存在性的API

interface PodTerminalProps {
  clusterName: string;
  namespace: string;
  podName: string;
  containerName: string;
}

/**
 * PodTerminal 组件用于连接到Kubernetes Pod的终端并提供交互式Shell
 * 
 * 该组件使用WebSocket建立与后端API的连接，并使用xterm.js提供终端UI
 * 包含错误处理和连接状态管理，确保在组件卸载时正确清理资源
 */
const PodTerminal: React.FC<PodTerminalProps> = ({
  clusterName,
  namespace,
  podName,
  containerName,
}) => {
  const terminalRef = useRef<HTMLDivElement>(null);
  const terminalInstance = useRef<Terminal | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'error' | 'closed' | 'checking'>('checking');
  const [errorMessage, setErrorMessage] = useState<string>('');
  const [reconnectCount, setReconnectCount] = useState(0);
  const [connectionTimeout, setConnectionTimeout] = useState<ReturnType<typeof setTimeout> | null>(null);
  const [isMounted, setIsMounted] = useState(false);
  const [podStatus, setPodStatus] = useState<'checking' | 'running' | 'notfound' | 'error'>('checking');
  const [containerList, setContainerList] = useState<string[]>([]);

  // 在组件挂载后设置状态，防止在卸载后的状态更新
  useEffect(() => {
    setIsMounted(true);
    return () => {
      setIsMounted(false);
      // 清除所有可能的超时
      if (connectionTimeout) {
        clearTimeout(connectionTimeout);
      }
    };
  }, [connectionTimeout]);

  // 检查Pod是否存在且运行中
  useEffect(() => {
    const checkPod = async () => {
      if (!isMounted) return;
      
      try {
        console.log('正在检查Pod状态:', clusterName, namespace, podName, containerName);
        const result = await checkPodExists(clusterName, namespace, podName);
        console.log('Pod状态检查结果:', result.data);
        
        if (result.data.code === 0 && result.data.data.exists) {
          // Pod存在，检查状态
          const podDetails = result.data.data.pod;
          console.log('Pod详情:', podDetails);
          
          const isRunning = podDetails?.status?.phase === 'Running';
          console.log('Pod是否运行中:', isRunning, podDetails?.status?.phase);
          
          // 获取容器列表，查看是否包含指定的容器
          const containerList = podDetails?.spec?.containers || [];
          const containerNames = containerList.map((c: any) => c.name);
          console.log('Pod中的容器列表:', containerNames);
          
          // 更新容器列表状态
          setContainerList(containerNames);
          
          // 如果containerName未指定或不在列表中，使用第一个容器
          let effectiveContainerName = containerName;
          if (!containerName && containerNames.length > 0) {
            effectiveContainerName = containerNames[0];
            console.log('未指定容器名，使用第一个容器:', effectiveContainerName);
            message.info(`未指定容器名，将使用 '${effectiveContainerName}'`);
          } else if (containerName && !containerNames.includes(containerName)) {
            // 尝试查找包含指定名称的容器
            const matchingContainer = containerNames.find((name: string) => name.includes(containerName));
            if (matchingContainer) {
              effectiveContainerName = matchingContainer;
              console.log('找到匹配的容器名:', effectiveContainerName);
              message.info(`找到匹配的容器名: '${effectiveContainerName}' (原请求: '${containerName}')`);
            } else {
              console.log('指定的容器不存在，使用第一个容器');
              if (containerNames.length > 0) {
                effectiveContainerName = containerNames[0];
                message.warning(`指定的容器 '${containerName}' 不存在，将使用 '${effectiveContainerName}'`);
              } else {
                message.error(`Pod 中没有可用的容器`);
              }
            }
          }
          
          // 检查容器是否就绪
          const containerStatuses = podDetails?.status?.containerStatuses || [];
          console.log('容器状态列表:', containerStatuses);
          
          // 查找匹配的容器状态
          const containerStatus = containerStatuses.find((cs: any) => cs.name === effectiveContainerName);
          const isReady = containerStatus?.ready === true;
          
          console.log('容器是否就绪:', isReady, effectiveContainerName, containerStatus);
          
          if (isRunning && containerStatuses.length > 0) {
            console.log('Pod和容器就绪，准备建立终端连接');
            // 先更新状态
            setPodStatus('running');
            // 使用 setTimeout 确保状态更新后再设置终端
            setTimeout(() => {
              if (isMounted) {
                setConnectionStatus('connecting');
                // 使用有效的容器名称
                setupTerminal(effectiveContainerName);
              }
            }, 0);
          } else {
            console.log('Pod或容器未就绪，无法建立连接');
            setPodStatus('notfound');
            setConnectionStatus('error');
            setErrorMessage(`容器未就绪: ${podName}/${effectiveContainerName || containerName} - ${podDetails?.status?.phase || 'Unknown'}`);
          }
        } else {
          console.log('Pod不存在');
          setPodStatus('notfound');
          setConnectionStatus('error');
          setErrorMessage(`Pod不存在: ${namespace}/${podName}`);
        }
      } catch (error: any) {
        console.error('检查Pod状态错误:', error);
        setPodStatus('error');
        setConnectionStatus('error');
        setErrorMessage(`无法检查Pod状态: ${error.message || '未知错误'}`);
      }
    };
    
    checkPod();
  }, [clusterName, namespace, podName, containerName, isMounted, reconnectCount]);

  // 执行终端fit调整大小，确保在终端和WebSocket都准备好后执行
  const safeResizeTerminal = () => {
    if (!isMounted || !terminalInstance.current || !fitAddonRef.current) return;
    
    try {
      fitAddonRef.current.fit();
      
      // 只有当WebSocket连接已建立并可用时才发送消息
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN && terminalInstance.current) {
        const dims = {
          cols: terminalInstance.current.cols,
          rows: terminalInstance.current.rows
        };
        wsRef.current.send(JSON.stringify({ type: 'resize', data: dims }));
      }
    } catch (error) {
      console.error('终端调整大小错误:', error);
    }
  };

  // 重连函数
  const reconnect = () => {
    // 清理现有连接
    cleanupResources();
    
    // 增加重连计数
    setReconnectCount(prev => prev + 1);
    
    // 重新初始化连接
    setConnectionStatus('checking');
    setPodStatus('checking');
    setErrorMessage('');
    
    // 延迟一点时间再重连，避免连续快速重连
    message.info('正在重新检查Pod状态...');
  };
  
  // 显示调试信息
  const showDebugInfo = () => {
    message.info(
      <div>
        <p>调试信息:</p>
        <p>集群: {clusterName}</p>
        <p>命名空间: {namespace}</p>
        <p>Pod名称: {podName}</p>
        <p>容器: {containerName}</p>
        <p>连接状态: {connectionStatus}</p>
        <p>Pod状态: {podStatus}</p>
        <p>可用容器: {containerList.join(', ')}</p>
        <p>WebSocket URL: {window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//{window.location.host}/api/clusters/{clusterName}/namespaces/{namespace}/pods/{podName}/exec?container={containerName}</p>
      </div>,
      10
    );
  };
  
  // 清理资源
  const cleanupResources = () => {
    try {
      if (wsRef.current) {
        if (wsRef.current.readyState === WebSocket.CONNECTING || 
            wsRef.current.readyState === WebSocket.OPEN) {
          wsRef.current.close();
        }
        wsRef.current = null;
      }
      
      if (terminalInstance.current) {
        terminalInstance.current.dispose();
        terminalInstance.current = null;
      }
      
      fitAddonRef.current = null;
      
      // 清除连接超时
      if (connectionTimeout) {
        clearTimeout(connectionTimeout);
        setConnectionTimeout(null);
      }
      
      // 移除窗口大小调整事件监听器
      window.removeEventListener('resize', safeResizeTerminal);
    } catch (error) {
      console.error('清理资源错误:', error);
    }
  };

  // 设置终端和WebSocket连接
  const setupTerminal = (effectiveContainerName?: string) => {
    if (!terminalRef.current || !isMounted) {
      console.log('终端设置条件不满足:', {
        terminalRefExists: !!terminalRef.current,
        isMounted
      });
      return;
    }
    
    console.log('开始设置终端和WebSocket连接');
    // 使用有效的容器名称或默认值
    const actualContainerName = effectiveContainerName || containerName;
    console.log('使用的容器名称:', actualContainerName);
    
    cleanupResources();

    // 创建终端实例
    const term = new Terminal({
      cursorBlink: true,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      fontSize: 14,
      theme: {
        background: '#1e1e1e'
      },
      allowTransparency: true
    });
    
    terminalInstance.current = term;
    console.log('终端实例已创建');
    
    // 创建并加载拟合插件
    const fitAddon = new FitAddon();
    fitAddonRef.current = fitAddon;
    
    try {
      term.loadAddon(fitAddon);
      
      // 打开终端
      term.open(terminalRef.current);
      term.writeln('正在连接到容器终端...');
      term.writeln(`连接到集群 ${clusterName}，命名空间 ${namespace}，Pod ${podName}，容器 ${actualContainerName}`);
      
      // 等待DOM更新后再调整大小
      setTimeout(() => {
        if (isMounted && fitAddonRef.current) {
          try {
            fitAddonRef.current.fit();
          } catch (error) {
            console.error('初始终端调整错误:', error);
          }
        }
      }, 100);
      
      // 创建WebSocket连接
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${protocol}//${window.location.host}/api/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/exec?container=${actualContainerName}`;
      
      console.log('尝试连接WebSocket:', wsUrl);
      console.log('当前环境:', {
        protocol: window.location.protocol,
        host: window.location.host,
        href: window.location.href
      });
      
      try {
        const ws = new WebSocket(wsUrl);
        console.log('WebSocket实例已创建');
        wsRef.current = ws;
        
        // 发送PING消息保持连接活跃
        const pingInterval = setInterval(() => {
          if (ws.readyState === WebSocket.OPEN) {
            try {
              ws.send(JSON.stringify({ type: 'ping' }));
            } catch (error) {
              console.error('发送ping消息失败:', error);
            }
          }
        }, 30000);
        
        // 设置连接超时
        const timeout = setTimeout(() => {
          if (wsRef.current && wsRef.current.readyState === WebSocket.CONNECTING) {
            console.error('WebSocket连接超时');
            term.writeln('\r\n连接超时，请检查网络或服务器状态。');
            setErrorMessage('连接超时 - 无法建立WebSocket连接');
            setConnectionStatus('error');
            ws.close();
            clearInterval(pingInterval);
          }
        }, 10000);
        
        setConnectionTimeout(timeout);
        
        // 连接成功时的处理
        ws.addEventListener('open', () => {
          if (!isMounted) return;
          
          console.log('WebSocket连接已打开');
          if (connectionTimeout) {
            clearTimeout(connectionTimeout);
            setConnectionTimeout(null);
          }
          
          term.writeln('\r\n连接成功! 您现在可以与容器终端交互。');
          setConnectionStatus('connected');
          
          try {
            if (ws && ws.readyState === WebSocket.OPEN) {
              const attachAddon = new AttachAddon(ws);
              term.loadAddon(attachAddon);
              setTimeout(safeResizeTerminal, 100);
            }
          } catch (error) {
            console.error('附加WebSocket错误:', error);
            term.writeln('\r\n附加WebSocket失败，请重试。');
            setErrorMessage('无法附加WebSocket到终端');
            clearInterval(pingInterval);
          }
        });
        
        // 连接关闭时的处理
        ws.addEventListener('close', (event) => {
          if (!isMounted) return;
          
          console.log('WebSocket连接关闭:', {
            code: event.code,
            reason: event.reason,
            wasClean: event.wasClean,
            timestamp: new Date().toISOString()
          });
          
          term.writeln('\r\n连接已关闭。');
          setConnectionStatus('closed');
          
          if (connectionTimeout) {
            clearTimeout(connectionTimeout);
            setConnectionTimeout(null);
          }
          clearInterval(pingInterval);
        });
        
        // 连接错误时的处理
        ws.addEventListener('error', (event) => {
          if (!isMounted) return;
          
          console.error('WebSocket错误:', {
            event,
            readyState: ws.readyState,
            url: wsUrl,
            timestamp: new Date().toISOString()
          });
          
          term.writeln('\r\n连接错误，请检查网络连接或服务器状态。');
          term.writeln('\r\n可能原因：');
          term.writeln('1. Pod或容器不存在或已终止');
          term.writeln('2. API服务器无法访问');
          term.writeln('3. 网络连接问题');
          term.writeln('4. WebSocket服务未启动或配置错误');
          term.writeln('\r\n请尝试刷新页面或点击重连按钮。');
          
          setConnectionStatus('error');
          setErrorMessage('WebSocket连接失败 - 请检查Pod和容器状态');
          
          if (connectionTimeout) {
            clearTimeout(connectionTimeout);
            setConnectionTimeout(null);
          }
          clearInterval(pingInterval);
        });
        
        // 添加窗口大小变化的监听器
        window.addEventListener('resize', safeResizeTerminal);
        
        // 清理函数
        return () => {
          clearInterval(pingInterval);
        };
        
      } catch (error: any) {
        console.error('创建WebSocket实例错误:', error.message || '未知错误');
        term.writeln(`\r\n创建WebSocket失败: ${error.message || '未知错误'}`);
        setErrorMessage(`无法创建WebSocket连接: ${error.message || '未知错误'}`);
        setConnectionStatus('error');
      }
    } catch (error: any) {
      console.error('终端设置错误:', error);
      setErrorMessage(`终端初始化失败 - ${error.message || '未知错误'}`);
      setConnectionStatus('error');
      
      // 清除超时
      if (connectionTimeout) {
        clearTimeout(connectionTimeout);
        setConnectionTimeout(null);
      }
    }
  };

  // 检查Pod中的容器列表
  const checkContainers = async () => {
    try {
      const result = await checkPodExists(clusterName, namespace, podName);
      if (result.data.code === 0 && result.data.data.exists) {
        const podDetails = result.data.data.pod;
        const containerList = podDetails?.spec?.containers || [];
        const containerNames = containerList.map((c: any) => c.name);
        
        setContainerList(containerNames);
        
        const containerListStr = containerNames.join('\n- ');
        message.info(
          <div>
            <p>Pod '{podName}' 中的容器列表:</p>
            <p>- {containerListStr}</p>
            <p>当前选择: {containerName || '未指定'}</p>
          </div>,
          10
        );
      } else {
        message.error(`Pod '${podName}' 不存在`);
      }
    } catch (error: any) {
      message.error(`检查容器列表失败: ${error.message || '未知错误'}`);
    }
  };

  // 渲染终端UI及错误状态
  return (
    <Card 
      title={
        <Space>
          容器终端
          {connectionStatus === 'connected' && <span style={{ color: '#52c41a' }}>(已连接)</span>}
          {connectionStatus === 'connecting' && <span style={{ color: '#faad14' }}>(连接中...)</span>}
          {connectionStatus === 'checking' && <span style={{ color: '#1890ff' }}>(检查Pod状态...)</span>}
          {(connectionStatus === 'error' || connectionStatus === 'closed') && 
            <span style={{ color: '#ff4d4f' }}>({connectionStatus === 'error' ? '连接错误' : '已断开'})</span>}
        </Space>
      }
      extra={
        <Space>
          <Button 
            icon={<BugOutlined />}
            onClick={checkContainers}
            size="small"
          >
            容器列表
          </Button>
          <Button 
            icon={<BugOutlined />}
            onClick={showDebugInfo}
            size="small"
          >
            调试信息
          </Button>
          <Button 
            type="primary" 
            icon={<ReloadOutlined />} 
            onClick={reconnect}
            disabled={connectionStatus === 'connecting' || connectionStatus === 'checking'}
          >
            重新连接
          </Button>
        </Space>
      }
      styles={{ 
        body: { 
          padding: 0, 
          height: '500px',
          display: 'flex',
          flexDirection: 'column'
        } 
      }}
    >
      {connectionStatus === 'error' && (
        <Alert
          message="连接错误"
          description={errorMessage || "无法连接到容器终端，请检查Pod和容器状态后重试。"}
          type="error"
          showIcon
          style={{ margin: '16px' }}
          action={
            <Button size="small" danger onClick={reconnect}>
              重试
            </Button>
          }
        />
      )}
      
      {(connectionStatus === 'connecting' || connectionStatus === 'checking') && (
        <div style={{ 
          display: 'flex', 
          justifyContent: 'center', 
          alignItems: 'center', 
          padding: '20px',
          backgroundColor: '#f0f2f5'
        }}>
          <Spin 
            tip={connectionStatus === 'checking' ? 
              "正在检查Pod状态..." : 
              "正在连接到容器终端..."} 
            size="large" 
          />
        </div>
      )}
      
      <div 
        ref={terminalRef} 
        style={{ 
          flexGrow: 1,
          background: '#1e1e1e',
          display: (connectionStatus === 'error' && !terminalInstance.current) || 
                  connectionStatus === 'checking' ? 'none' : 'block'
        }} 
      />
    </Card>
  );
};

export default PodTerminal;